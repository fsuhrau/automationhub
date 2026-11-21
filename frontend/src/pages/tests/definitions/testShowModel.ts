// The test page without React: the results of a run in words and the picked
// Unity tests grouped by class.
import {ITestRunListEntry} from '../../../types/test.run.list';

type Counted = Pick<ITestRunListEntry, 'passed' | 'failed' | 'unstable' | 'open'>;

// runResults says how a run went: "3 passed · 1 failed"; empty without counts
export const runResults = (r: Counted): string =>
    [r.passed && `${r.passed} passed`, r.failed && `${r.failed} failed`, r.unstable && `${r.unstable} unstable`, r.open && `${r.open} open`]
        .filter(Boolean).join(' · ');

interface Named {
    class: string,
    assembly: string,
    method: string,
}

// methodsByClass groups the methods by their class (or assembly), in the
// order the classes first appear
export const methodsByClass = (functions: Named[]): [string, string[]][] => {
    const groups = new Map<string, string[]>();
    functions.forEach(f => {
        const key = f.class || f.assembly || '';
        groups.set(key, [...(groups.get(key) ?? []), f.method]);
    });
    return Array.from(groups);
};
