// One labelled fact on the bundle page. Empty facts are not shown.

export interface FactSpec {
    label: string,
    text?: string,
    // shown as chips instead of text
    chips?: string[],
    // chips shown before "+n more"
    max?: number,
    // a date shown with a mark when it has passed or is near
    expiry?: string,
    // adds a button that copies text
    copy?: boolean,
    monospace?: boolean,
}
