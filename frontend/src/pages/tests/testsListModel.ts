// The tests list without React: how a test is doing (from its latest run),
// the filters by that, search, and one line about what a test runs.
import {IRunSummary, RunSummaryStatus} from '../../types/test';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {TestType} from '../../types/test.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';

export type Tone = 'success' | 'error' | 'warning' | 'info' | 'default';

export const STATUS: Record<RunSummaryStatus, { label: string, tone: Tone }> = {
    running: {label: 'Running', tone: 'info'},
    passed: {label: 'Passed', tone: 'success'},
    failed: {label: 'Failed', tone: 'error'},
    unstable: {label: 'Unstable', tone: 'warning'},
    interrupted: {label: 'Interrupted', tone: 'default'},
    not_run: {label: 'Did not run', tone: 'error'},
    empty: {label: 'No results', tone: 'default'},
};

export type Health = 'failing' | 'passing' | 'running' | 'never';

// health is how a test is doing: its latest run decides; a running test
// counts as running even when an earlier run failed
export const health = (recent: IRunSummary[] | undefined): Health => {
    const last = recent?.[0];
    if (!last) return 'never';
    if (last.status === 'running') return 'running';
    if (last.status === 'passed') return 'passing';
    if (last.status === 'empty' || last.status === 'interrupted') {
        // nothing to judge by; the run before it decides
        const before = recent!.slice(1).find(r => r.status !== 'empty' && r.status !== 'interrupted' && r.status !== 'running');
        return before?.status === 'passed' ? 'passing' : before ? 'failing' : 'never';
    }
    return 'failing';
};

export type HealthFilter = Health | 'all';

export interface TestLike {
    id: number,
    name: string,
    recentRuns?: IRunSummary[],
    testConfig: TestConfigLike,
}

export const countHealth = (tests: TestLike[]): Record<HealthFilter, number> => {
    const counts: Record<HealthFilter, number> = {all: tests.length, failing: 0, passing: 0, running: 0, never: 0};
    tests.forEach(t => counts[health(t.recentRuns)]++);
    return counts;
};

export const filterTests = <T extends TestLike>(tests: T[], filter: HealthFilter, text: string): T[] => {
    const needle = text.trim().toLowerCase();
    return tests.filter(t => (filter === 'all' || health(t.recentRuns) === filter)
        && (needle === '' || t.name.toLowerCase().includes(needle)));
};

export interface TestConfigLike {
    type: TestType,
    executionType: TestExecutionType,
    allDevices: boolean,
    devices?: unknown[] | null,
    unity?: { testCategoryType: UnityTestCategory, testFunctions?: unknown[] | null, categories?: string } | null,
}

const plural = (n: number, one: string, many: string) => `${n} ${n === 1 ? one : many}`;

// describeConfig says in one line what a test runs and where:
// "Unity · all tests · on all devices · one device after the other"
export const describeConfig = (c: TestConfigLike): string => {
    const parts: string[] = [TestType[c.type] ?? 'Test'];
    if (c.type === TestType.Unity && c.unity) {
        switch (c.unity.testCategoryType) {
            case UnityTestCategory.RunAllTests:
                parts.push('all tests');
                break;
            case UnityTestCategory.RunAllOfCategory:
                parts.push(c.unity.categories ? `category ${c.unity.categories}` : 'a category');
                break;
            default:
                parts.push(plural(c.unity.testFunctions?.length ?? 0, 'test', 'tests'));
        }
    }
    parts.push(c.allDevices ? 'on all devices' : `on ${plural(c.devices?.length ?? 0, 'device', 'devices')}`);
    parts.push(c.executionType === TestExecutionType.Simultaneously ? 'devices at the same time' : 'one device after the other');
    return parts.join(' · ');
};

export interface RunLike {
    createdAt: Date | string,
    endedAt?: Date | string | null,
    status?: RunSummaryStatus,
}

export interface RunStats {
    // runs with a verdict (not running, not interrupted or empty)
    judged: number,
    passed: number,
    lastPassedAt: Date | string | null,
    // the middle duration of the passed runs, in ms
    typicalMs: number | null,
}

// runStats sums up runs (newest first) for the test page
export const runStats = (runs: RunLike[]): RunStats => {
    const judged = runs.filter(r => r.status === 'passed' || r.status === 'failed' || r.status === 'unstable' || r.status === 'not_run');
    const passed = judged.filter(r => r.status === 'passed');
    const durations = passed.filter(r => r.endedAt)
        .map(r => new Date(r.endedAt!).valueOf() - new Date(r.createdAt).valueOf())
        .filter(ms => ms >= 0).sort((a, b) => a - b);
    return {
        judged: judged.length,
        passed: passed.length,
        lastPassedAt: passed[0]?.createdAt ?? null,
        typicalMs: durations.length ? durations[Math.floor((durations.length - 1) / 2)] : null,
    };
};
