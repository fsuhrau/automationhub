import {expect, it} from 'vitest';
import {archiveText, deleteText} from './bundleDialogsModel';

it('says how big the file is and how many runs used it', () => {
    expect(deleteText({size: 2048, runCount: 1})).toBe(
        'The file (2.00KB) is removed from the master. 1 test run used it; their results stay, but they can\'t be run again with this bundle.');
    expect(deleteText({size: 0, runCount: 3})).toContain('3 test runs used it');
});

it('says no run used the bundle, or nothing when the count is unknown', () => {
    expect(deleteText({size: 0, runCount: 0})).toBe('The file is removed from the master. No test run used it.');
    expect(deleteText({size: 0})).toBe('The file is removed from the master.');
});

it('says what an archive keeps', () => {
    expect(archiveText({size: 1024})).toContain('The file (1.00KB) is removed from the master. Everything read from it');
    expect(archiveText({size: 1024})).toContain('uploading the same file again brings it back');
});
