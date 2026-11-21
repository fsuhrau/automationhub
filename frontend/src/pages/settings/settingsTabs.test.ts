import {expect, it} from 'vitest';
import {parseTab, withTab} from './settingsTabs';

it.each([['apps', 'apps'], ['nodes', 'nodes'], ['nope', 'project'], [null, 'project']])('parseTab: %s gives %s', (value, tab) => {
    expect(parseTab(value)).toBe(tab);
});

it('withTab: sets the tab and keeps other parameters', () => {
    expect(withTab(new URLSearchParams('x=1'), 'tokens').toString()).toBe('x=1&tab=tokens');
});

it('withTab: removes the parameter for the first tab', () => {
    expect(withTab(new URLSearchParams('tab=apps'), 'project').toString()).toBe('');
});
