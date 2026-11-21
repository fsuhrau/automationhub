// The words of the dialogs that archive or delete a bundle.
import {byteFormat} from '../tests/value_formatter';

interface Described {
    size: number,
    runCount?: number,
}

const fileSize = (b: Described): string => b.size ? ` (${byteFormat(b.size)})` : '';

const runsUsed = (count: number): string =>
    ` ${count} test ${count === 1 ? 'run used' : 'runs used'} it; their results stay, but they can't be run again with this bundle.`;

// the run count is unknown on older masters: say nothing rather than "none"
const runsText = (count: number | undefined): string => {
    if (count === undefined) return '';
    return count > 0 ? runsUsed(count) : ' No test run used it.';
};

export const deleteText = (b: Described): string => `The file${fileSize(b)} is removed from the master.${runsText(b.runCount)}`;

export const archiveText = (b: Described): string =>
    `The file${fileSize(b)} is removed from the master. Everything read from it, its tags and the results of its test runs stay, `
    + 'so runs with it can still be compared. It can\'t be installed or run anymore; uploading the same file again brings it back.';
