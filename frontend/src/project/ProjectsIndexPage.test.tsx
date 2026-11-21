import {afterEach, expect, it, vi} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ProjectsIndexPage from './ProjectsIndexPage';
import {app, apiError, createFakeApi, Handler, Opened, project, renderOutsideProject} from '../pages/settings/testKit';

vi.mock('../components/TopBar', () => ({default: () => <header>top bar</header>}));

afterEach(() => localStorage.clear());

const alpha = project({id: 1, identifier: 'alpha', name: 'Alpha', apps: [app({id: 1, name: 'Game Android'})]});
const bravo = project({id: 2, identifier: 'bravo', name: 'Bravo', apps: ['a', 'b', 'c', 'd', 'e', 'f'].map((n, i) => app({id: 10 + i, name: `App ${n}`}))});

const open = (projects = [alpha, bravo], handlers: Record<string, Handler> = {}): Opened => {
    const api = createFakeApi({'GET /projects': () => projects, ...handlers});
    renderOutsideProject(<ProjectsIndexPage/>, api);
    return {api, user: userEvent.setup()};
};

it('ProjectsIndexPage: shows a loading bar, then the projects with their apps', async () => {
    open([alpha, bravo, project({id: 3, identifier: 'charlie', name: 'Charlie'})]);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
    expect(await screen.findByRole('heading', {name: 'Alpha'})).toBeInTheDocument();
    expect(screen.queryByRole('progressbar')).toBeNull();
    expect(screen.getByText('Game Android')).toBeInTheDocument();
    expect(screen.getByText('and 2 more')).toBeInTheDocument();
    expect(screen.getByText('No apps yet', {exact: false})).toBeInTheDocument();
});

it('ProjectsIndexPage: lists the last opened project first and marks it', async () => {
    localStorage.setItem('lastProject', 'bravo');
    open();
    const names = (await screen.findAllByRole('heading', {level: 2})).map(h => h.textContent);
    expect(names).toEqual(['Bravo', 'Alpha']);
    expect(screen.getByText('last opened')).toBeInTheDocument();
});

it('ProjectsIndexPage: links to the project and to its settings', async () => {
    open();
    await screen.findByRole('heading', {name: 'Alpha'});
    expect(screen.getByRole('link', {name: /^Alpha/})).toHaveAttribute('href', '/project/alpha');
    expect(screen.getByRole('link', {name: 'Settings of Alpha'})).toHaveAttribute('href', '/project/alpha/settings');
});

it('ProjectsIndexPage: offers a search for many projects and filters by project or app name', async () => {
    const many = Array.from({length: 7}, (_, i) => project({id: i, identifier: `p${i}`, name: `Proj ${i}`, apps: [app({name: `Tool ${i}`})]}));
    const {user} = open(many);
    await screen.findByRole('heading', {name: 'Proj 0'});
    await user.type(screen.getByRole('textbox', {name: 'Find a project'}), 'tool 3');
    expect(screen.getAllByRole('heading', {level: 2}).map(h => h.textContent)).toEqual(['Proj 3']);
    await user.clear(screen.getByRole('textbox', {name: 'Find a project'}));
    await user.type(screen.getByRole('textbox', {name: 'Find a project'}), 'zzz');
    expect(screen.getByText('No project matches.')).toBeInTheDocument();
});

it('ProjectsIndexPage: has no search for a short list', async () => {
    open();
    await screen.findByRole('heading', {name: 'Alpha'});
    expect(screen.queryByRole('textbox', {name: 'Find a project'})).toBeNull();
});

it('ProjectsIndexPage without projects, and errors: invites to create the first project', async () => {
    const {user} = open([]);
    expect(await screen.findByText('No projects yet')).toBeInTheDocument();
    await user.click(screen.getAllByRole('button', {name: 'New project'})[1]);
    expect(await screen.findByRole('dialog', {name: 'New project'})).toBeInTheDocument();
});

it('ProjectsIndexPage without projects, and errors: shows the error and loads again on retry', async () => {
    let fail = true;
    const {api, user} = open([alpha], {'GET /projects': () => {
        if (fail) throw apiError('master is down');
        return [alpha];
    }});
    expect(await screen.findByText('master is down')).toBeInTheDocument();
    fail = false;
    await user.click(screen.getByRole('button', {name: 'Retry'}));
    expect(await screen.findByRole('heading', {name: 'Alpha'})).toBeInTheDocument();
    expect(api.bodies('GET /projects')).toHaveLength(2);
});

const openDialog = async (handlers: Record<string, Handler> = {}): Promise<Opened & { dialog: HTMLElement }> => {
    const result = open([alpha], handlers);
    await screen.findByRole('heading', {name: 'Alpha'});
    await result.user.click(screen.getByRole('button', {name: 'New project'}));
    return {...result, dialog: await screen.findByRole('dialog', {name: 'New project'})};
};

it('ProjectsIndexPage > new project: previews the identifier, creates the project and opens its settings to add an app', async () => {
    const created = project({id: 3, identifier: 'heroes-of-history', name: 'Heroes of History'});
    const {api, user, dialog} = await openDialog({'POST /project': () => created});
    await user.type(within(dialog).getByLabelText('Name'), ' Heroes of History ');
    expect(within(dialog).getByText('heroes-of-history')).toBeInTheDocument();
    await user.click(within(dialog).getByRole('button', {name: 'Create project'}));
    expect(await screen.findByText('at /project/heroes-of-history/settings?tab=apps')).toBeInTheDocument();
    expect(api.bodies('POST /project')).toEqual([{name: 'Heroes of History'}]);
});

it('ProjectsIndexPage > new project: asks for a name first', async () => {
    const {api, user, dialog} = await openDialog();
    await user.click(within(dialog).getByRole('button', {name: 'Create project'}));
    expect(within(dialog).getByText('A name is needed.')).toBeInTheDocument();
    expect(api.bodies('POST /project')).toEqual([]);
});

it('ProjectsIndexPage > new project: shows the master\'s error in the dialog, and can be cancelled', async () => {
    const {user, dialog} = await openDialog({'POST /project': () => {
        throw apiError('quota reached');
    }});
    await user.type(within(dialog).getByLabelText('Name'), 'X');
    await user.click(within(dialog).getByRole('button', {name: 'Create project'}));
    expect(await within(dialog).findByText('quota reached')).toBeInTheDocument();
    await user.click(within(dialog).getByRole('button', {name: 'Cancel'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
});
