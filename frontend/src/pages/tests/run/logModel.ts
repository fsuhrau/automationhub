// The log side of a test's protocol: runtimes, the runner's setup and
// teardown merged into the test's log, summaries, filters and the text download.
// Kept free of React so it can be checked on its own.
import IProtocolEntryData from '../../../types/protocol.entry';
import ITestProtocolData from '../../../types/test.protocol';

// formatRuntime shows seconds since the test started as m:ss.mmm (h:mm:ss
// past an hour), the clock the log is read by. The runner's setup happens
// before the test starts, so its times are negative: -0:02.345.
export const formatRuntime = (seconds: number | null | undefined): string => {
    const v = seconds ?? 0;
    // a sign only for what shows as more than zero
    const sign = v <= -0.0005 ? '-' : '';
    const s = Math.abs(v);
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const sec = s % 60;
    if (h > 0) {
        return `${sign}${h}:${String(m).padStart(2, '0')}:${String(Math.floor(sec)).padStart(2, '0')}`;
    }
    return `${sign}${m}:${sec.toFixed(3).padStart(6, '0')}`;
};

export const isError = (e: IProtocolEntryData): boolean => e.level === 'error';
export const isScreen = (e: IProtocolEntryData): boolean => e.source === 'screen' && !!e.data;
export const isStep = (e: IProtocolEntryData): boolean => e.source === 'step';

// sourceName labels the log sources; entries without one come from the runner
export const sourceName = (source: string): string => source === '' ? 'runner' : source;

// Phase tells the runner's entries apart in a test's log: what the hub and
// the app logged before the test method started, and after it ended.
export type Phase = 'setup' | 'teardown';

export interface LogEntry extends IProtocolEntryData {
    // set on the runner's entries only
    phase?: Phase,
}

const seconds = (d: Date | string | null | undefined): number | null => d ? new Date(d).valueOf() / 1000 : null;

// mergeRunnerLog puts the runner's entries into the test's log, on the
// test's clock: setup before 0:00, teardown after the test's last entry.
// The runner writes nothing while the test runs, the test has the log then.
export const mergeRunnerLog = (test: ITestProtocolData, entries: IProtocolEntryData[], runner: ITestProtocolData | null): LogEntry[] => {
    const runnerEntries = runner?.entries ?? [];
    if (!runner || runnerEntries.length === 0) {
        return entries;
    }
    const testStart = seconds(test.startedAt);
    const runnerStart = seconds(runner.startedAt);
    const offset = testStart !== null && runnerStart !== null ? testStart - runnerStart : 0;
    const shifted: LogEntry[] = runnerEntries.map(e => {
        const runtime = e.runtime - offset;
        return {...e, runtime, phase: runtime < 0 ? 'setup' : 'teardown'};
    });
    // stable: on equal times the test's entries come first, each in its order
    return [...entries, ...shifted]
        .map((e, i) => ({e, i}))
        .sort((a, b) => a.e.runtime - b.e.runtime || a.i - b.i)
        .map(x => x.e);
};

export interface LogSummary<T extends IProtocolEntryData = IProtocolEntryData> {
    errors: T[],
    steps: T[],
    screens: T[],
    // entries per source, in the order sources first appear
    sources: { source: string, count: number }[],
}

export const summarize = <T extends IProtocolEntryData>(entries: T[]): LogSummary<T> => {
    const counts = new Map<string, number>();
    entries.forEach(e => counts.set(sourceName(e.source), (counts.get(sourceName(e.source)) ?? 0) + 1));
    return {
        errors: entries.filter(isError),
        steps: entries.filter(isStep),
        screens: entries.filter(isScreen),
        sources: Array.from(counts, ([source, count]) => ({source, count})),
    };
};

export interface LogFilter {
    // sources and runner phases switched off
    hidden: Set<string>,
    errorsOnly: boolean,
    text: string,
}

export const filterLog = (entries: LogEntry[], f: LogFilter): LogEntry[] => {
    const needle = f.text.trim().toLowerCase();
    return entries.filter(e => (!f.errorsOnly || isError(e))
        && !f.hidden.has(sourceName(e.source))
        && !(e.phase && f.hidden.has(e.phase))
        && (needle === '' || (e.message ?? '').toLowerCase().includes(needle)));
};

// headline is the first line of a message, the rest is its detail (stack
// traces, JSON)
export const splitMessage = (message: string): { headline: string, detail: string } => {
    const text = (message ?? '').replace(/\s+$/, '');
    const nl = text.indexOf('\n');
    return nl < 0 ? {headline: text, detail: ''} : {headline: text.substring(0, nl), detail: text.substring(nl + 1)};
};

const textLine = (e: LogEntry): string => {
    const source = `${e.phase ? `${e.phase}/` : ''}${sourceName(e.source)}`;
    const message = (e.message ?? '').replace(/\n/g, '\n\t\t\t');
    return `${formatRuntime(e.runtime)}\t${source}\t${e.level}\t${message}${e.data ? `\t[${e.data}]` : ''}`;
};

// logAsText is the log as a plain text file for download
export const logAsText = (testName: string, entries: LogEntry[]): string =>
    [`# ${testName}`, ...entries.map(textLine)].join('\n') + '\n';
