import {afterEach, expect, it, vi} from 'vitest';
import {lastProject, ProjectLike, matchesProject, projectSlug, rememberProject, sortProjects, visibleProjects} from './projectsModel';

const p = (identifier: string, name: string, apps: { name: string }[] = []): ProjectLike => ({identifier, name, apps});

it.each([
    ['Heroes of History', 'heroes-of-history'],
    ['  --A!!B-- ', 'a-b'],
    ['!!!', 'project'],
    ['x'.repeat(60), 'x'.repeat(48)],
    [`${'a'.repeat(47)} b`, 'a'.repeat(47)],
])('projectSlug: %s gives %s', (name, slug) => {
    expect(projectSlug(name)).toBe(slug);
});

it('matchesProject: matches everything for an empty search', () => {
    expect(matchesProject(p('a', 'A'), '  ')).toBe(true);
});

it.each([['hero', true], ['ZZ', false], ['alpha', true], ['pro', true]])('matchesProject: searching %s gives %s', (text, found) => {
    expect(matchesProject(p('pro-1', 'Heroes', [{name: 'Alpha'}]), text)).toBe(found);
});

it('matchesProject: tolerates projects without apps', () => {
    expect(matchesProject({identifier: 'a', name: 'A', apps: null}, 'x')).toBe(false);
    expect(matchesProject({identifier: 'a', name: 'A'}, 'x')).toBe(false);
});

const list = [p('b', 'Bravo'), p('c', 'Charlie'), p('a', 'Alpha')];

it('sortProjects and visibleProjects: puts the last opened first, then by name', () => {
    expect(sortProjects(list, 'c').map(x => x.identifier)).toEqual(['c', 'a', 'b']);
    expect(sortProjects(list, null).map(x => x.identifier)).toEqual(['a', 'b', 'c']);
});

it('sortProjects and visibleProjects: filters the sorted list by the search', () => {
    expect(visibleProjects(list, 'c', 'a').map(x => x.identifier)).toEqual(['c', 'a', 'b']);
    expect(visibleProjects(list, null, 'bra').map(x => x.identifier)).toEqual(['b']);
});

afterEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
});

it('remembering the last project: stores and reads it', () => {
    expect(lastProject()).toBeNull();
    rememberProject('p1');
    expect(lastProject()).toBe('p1');
});

it('remembering the last project: survives a storage that throws', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
        throw new Error('denied');
    });
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
        throw new Error('denied');
    });
    expect(() => rememberProject('p1')).not.toThrow();
    expect(lastProject()).toBeNull();
});
