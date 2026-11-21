// hideBelow hides a table column on screens narrower than the breakpoint
export const hideBelow = (breakpoint: 'sm' | 'md'): object => ({display: {xs: 'none', [breakpoint]: 'table-cell'}});
