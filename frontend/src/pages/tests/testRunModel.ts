// The data side of the test run page: merging live protocol updates,
// grouping them into rows, counting results, and the run's overall status.
// Kept free of React so it can be checked on its own.
import ITestProtocolData from '../../types/test.protocol';
import {TestResultState} from '../../types/test.result.state.enum';

// the run's lifecycle as the master reports it; undefined for masters that
// predate the field
export type RunLifecycle = 'running' | 'finished' | 'not_running' | undefined;

export const isTopLevel = (p: ITestProtocolData): boolean =>
    p.parentTestProtocolId === null || p.parentTestProtocolId === undefined || p.parentTestProtocolId === 0;

// mergeProtocol upserts a protocol update into the flat list. Updates are
// published concurrently, so a late "open" update must not overwrite a
// finished protocol, and updates usually don't carry the preloaded device.
export const mergeProtocol = (protocols: ITestProtocolData[], update: ITestProtocolData | null): ITestProtocolData[] => {
    if (!update || update.id === undefined || update.id === null) {
        return protocols;
    }
    const index = protocols.findIndex(p => p.id === update.id);
    if (index < 0) {
        return [...protocols, update];
    }
    const existing = protocols[index];
    const merged: ITestProtocolData = {...existing, ...update, device: update.device ?? existing.device};
    if (existing.endedAt && !update.endedAt) {
        merged.endedAt = existing.endedAt;
        merged.testResult = existing.testResult;
    }
    const next = [...protocols];
    next[index] = merged;
    return next;
};

// splitTestName turns "Namespace.Class/Method" into the method and its suite.
export const splitTestName = (testName: string): { name: string, suite: string } => {
    const slash = testName.lastIndexOf('/');
    if (slash < 0) {
        return {name: testName, suite: ''};
    }
    return {name: testName.substring(slash + 1), suite: testName.substring(0, slash)};
};

export interface TestRow {
    key: number,
    // the protocol the row opens
    protocol: ITestProtocolData,
    // the runner's protocol, when the row is a test that the runner wrapped
    runner: ITestProtocolData | null,
    name: string,
    suite: string,
    result: TestResultState,
    startedAt: Date,
    endedAt?: Date,
    children: TestRow[],
}

const leafRow = (p: ITestProtocolData, runner: ITestProtocolData | null = null): TestRow => ({
    key: p.id as number,
    protocol: p,
    runner,
    ...splitTestName(p.testName),
    result: p.testResult,
    // the runner times the test as the run log does; the test's own start is
    // taken when the app connects and can be far off
    startedAt: (runner ?? p).startedAt,
    endedAt: (runner ?? p).endedAt,
    children: [],
});

// buildRows nests child protocols under their parent. A runner protocol with
// a single child of the same test is one row: the row opens the test's own
// protocol (its steps, screenshots and performance), the runner's is a side
// link. Children whose parent hasn't arrived yet show up with it.
export const buildRows = (protocols: ITestProtocolData[]): TestRow[] => {
    const children = new Map<number, ITestProtocolData[]>();
    protocols.filter(p => !isTopLevel(p)).forEach(p => {
        const list = children.get(p.parentTestProtocolId as number) ?? [];
        list.push(p);
        children.set(p.parentTestProtocolId as number, list);
    });
    return protocols.filter(isTopLevel).map(parent => {
        const kids = children.get(parent.id as number) ?? [];
        if (kids.length === 1 && kids[0].testName === parent.testName) {
            return leafRow(kids[0], parent);
        }
        return {...leafRow(parent), children: kids.map(k => leafRow(k))};
    });
};

// leaves are the executed tests: the children of a row, or the row itself
export const leaves = (rows: TestRow[]): TestRow[] =>
    rows.flatMap(r => r.children.length > 0 ? r.children : [r]);

export type ResultCounts = Record<TestResultState, number>;

export const countResults = (rows: TestRow[]): ResultCounts => {
    const counts: ResultCounts = {
        [TestResultState.TestResultOpen]: 0,
        [TestResultState.TestResultUnstable]: 0,
        [TestResultState.TestResultFailed]: 0,
        [TestResultState.TestResultSuccess]: 0,
    };
    leaves(rows).forEach(r => counts[r.result]++);
    return counts;
};

export type ResultFilter = TestResultState | 'all';

// filterRows keeps a row when it or one of its children matches the result
// and the text, and filters its children the same way.
export const filterRows = (rows: TestRow[], result: ResultFilter, text: string): TestRow[] => {
    const needle = text.trim().toLowerCase();
    const matches = (r: TestRow) => (result === 'all' || r.result === result)
        && (needle === '' || r.name.toLowerCase().includes(needle) || r.suite.toLowerCase().includes(needle));
    return rows.flatMap(r => {
        if (r.children.length === 0) {
            return matches(r) ? [r] : [];
        }
        const kids = r.children.filter(matches);
        return kids.length > 0 ? [{...r, children: kids}] : [];
    });
};

export type RunStatus = 'running' | 'failed' | 'unstable' | 'interrupted' | 'success' | 'not_run' | 'empty';

// runStatus is the one word for the whole run. A run that stopped while tests
// were still open (cancelled, crashed, or the master restarted) is
// "interrupted" unless something already failed; one without any result but
// with errors in its log never got to run a test.
export const runStatus = (lifecycle: RunLifecycle, counts: ResultCounts, logErrors = 0): RunStatus => {
    if (lifecycle === 'running') return 'running';
    const total = Object.values(counts).reduce((a, b) => a + b, 0);
    if (total === 0) return logErrors > 0 ? 'not_run' : 'empty';
    if (counts[TestResultState.TestResultFailed] > 0) return 'failed';
    if (counts[TestResultState.TestResultUnstable] > 0) return 'unstable';
    if (counts[TestResultState.TestResultOpen] > 0) return 'interrupted';
    if (counts[TestResultState.TestResultSuccess] > 0) return 'success';
    return 'empty';
};

// runEnd is when the run's last test or log line ended; null while running
export const runEnd = (lifecycle: RunLifecycle, protocols: ITestProtocolData[], logTimes: Date[]): Date | null => {
    if (lifecycle === 'running') return null;
    const times = [...protocols.map(p => p.endedAt).filter((d): d is Date => !!d), ...logTimes]
        .map(d => new Date(d).valueOf());
    return times.length > 0 ? new Date(Math.max(...times)) : null;
};

// parseParameters splits "A=1;B=2" into pairs; a name without "=" has no
// value. Runs store them one per line, the run request separates them by ";".
export const parseParameters = (parameter: string | null | undefined): { name: string, value: string }[] =>
    (parameter ?? '').split(/[;\r\n]+/).map(p => p.trim()).filter(p => p.length > 0).map(p => {
        const eq = p.indexOf('=');
        return eq < 0 ? {name: p, value: ''} : {name: p.substring(0, eq), value: p.substring(eq + 1)};
    });

// formatDuration gives "7m 44s" style durations from milliseconds
export const formatDuration = (ms: number): string => {
    const total = Math.max(0, Math.floor(ms / 1000));
    const h = Math.floor(total / 3600);
    const m = Math.floor((total % 3600) / 60);
    const s = total % 60;
    if (h > 0) return `${h}h ${m}m`;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
};

export const between = (start: Date, end: Date | null | undefined): number | null =>
    end ? new Date(end).valueOf() - new Date(start).valueOf() : null;
