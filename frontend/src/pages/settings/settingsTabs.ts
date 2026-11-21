// The tabs of the settings page; the active one is in the URL.

export const SETTINGS_TABS = [
    {key: 'project', label: 'Project'},
    {key: 'apps', label: 'Apps'},
    {key: 'tokens', label: 'Access tokens'},
    {key: 'nodes', label: 'Nodes'},
    {key: 'people', label: 'People'},
] as const;

export type TabKey = typeof SETTINGS_TABS[number]['key'];

const DEFAULT_TAB: TabKey = 'project';

// parseTab is the tab a URL asks for, the first one when it asks for none
// or an unknown one
export const parseTab = (value: string | null): TabKey =>
    SETTINGS_TABS.find(t => t.key === value)?.key ?? DEFAULT_TAB;

// withTab sets the tab in the query; the first tab needs no parameter
export const withTab = (query: URLSearchParams, tab: TabKey): URLSearchParams => {
    const next = new URLSearchParams(query);
    if (tab === DEFAULT_TAB) next.delete('tab'); else next.set('tab', tab);
    return next;
};
