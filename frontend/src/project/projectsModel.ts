// The projects page without React: the identifier a name gets (the master's
// rule, for a preview), finding projects, and the last opened one.

// projectSlug mirrors the master: lower case letters, digits and single
// dashes, at most 48 characters, "project" when nothing is left. The master
// adds -2, -3 … when the identifier is taken.
export const projectSlug = (name: string): string => {
    let slug = '';
    let dash = false;
    for (const ch of name.toLowerCase()) {
        if ((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
            slug += ch;
            dash = false;
        } else if (slug.length > 0 && !dash) {
            slug += '-';
            dash = true;
        }
    }
    slug = slug.replace(/-$/, '');
    if (slug.length > 48) slug = slug.slice(0, 48).replace(/-$/, '');
    return slug || 'project';
};

export interface ProjectLike {
    identifier: string,
    name: string,
    apps?: { name: string }[] | null,
}

export const matchesProject = (p: ProjectLike, text: string): boolean => {
    const needle = text.trim().toLowerCase();
    return needle === '' || p.name.toLowerCase().includes(needle) || p.identifier.includes(needle)
        || (p.apps ?? []).some(a => a.name.toLowerCase().includes(needle));
};

// sortProjects puts the last opened project first, the rest by name
export const sortProjects = <T extends ProjectLike>(projects: T[], last: string | null): T[] =>
    [...projects].sort((a, b) => (a.identifier === last ? -1 : b.identifier === last ? 1 : a.name.localeCompare(b.name)));

const LAST_KEY = 'lastProject';

export const rememberProject = (identifier: string) => {
    try {
        localStorage.setItem(LAST_KEY, identifier);
    } catch {
        // a convenience only
    }
};

export const lastProject = (): string | null => {
    try {
        return localStorage.getItem(LAST_KEY);
    } catch {
        return null;
    }
};
