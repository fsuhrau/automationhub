import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import SettingsPage from './SettingsPage';
import {apiError, app, createFakeApi, Handler, Opened, parameter, renderInProject, User} from './testKit';

const android = app({parameter: [parameter(), parameter({id: 11, name: 'Mode', type: {type: 'option', options: ['a', 'b'], defaultValue: 'a'}})]});

const open = async (handlers: Record<string, Handler> = {}): Promise<Opened> => {
    const api = createFakeApi({'GET /p1/apps': () => [android], ...handlers});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=apps'});
    await screen.findByRole('table', {name: 'Parameters'});
    return {api, user: userEvent.setup()};
};

const removeChip = async (user: User, within_: HTMLElement, label: string): Promise<void> => {
    within(within_).getByRole('button', {name: label}).focus();
    await user.keyboard('{Backspace}');
};

const dialog = (name: RegExp | string): Promise<HTMLElement> => screen.findByRole('dialog', {name});

it('Settings > Apps > parameters: adds a text parameter', async () => {
    const {api, user} = await open({'POST /p1/app/1/parameter': () => parameter({id: 12, name: 'Level', type: {type: 'string', defaultValue: '3'}})});
    await user.click(screen.getByRole('button', {name: 'Add parameter'}));
    const d = await dialog('Add a parameter');
    await user.type(within(d).getByLabelText('Name'), 'Level');
    await user.type(within(d).getByLabelText('Default'), '3');
    await user.click(within(d).getByRole('button', {name: 'Add parameter'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(api.bodies('POST /p1/app/1/parameter')).toEqual([{name: 'Level', type: {type: 'string', defaultValue: '3'}}]);
    expect(within(screen.getByRole('table', {name: 'Parameters'})).getByText('Level')).toBeInTheDocument();
});

it('Settings > Apps > parameters: adds a choice parameter from typed options and a picked default', async () => {
    const region = parameter({id: 12, name: 'Region', type: {type: 'option', options: ['eu', 'us'], defaultValue: 'us'}});
    const {api, user} = await open({'POST /p1/app/1/parameter': () => region});
    await user.click(screen.getByRole('button', {name: 'Add parameter'}));
    const d = await dialog('Add a parameter');
    await user.type(within(d).getByLabelText('Name'), 'Region');
    await user.click(within(d).getByRole('radio', {name: 'Choice of options'}));
    await user.type(within(d).getByLabelText('Option'), 'eu,us,junk{Enter}');
    await removeChip(user, d, 'junk');
    await user.click(within(d).getByRole('combobox', {name: /Default/}));
    await user.click(screen.getByRole('option', {name: 'us'}));
    await user.click(within(d).getByRole('button', {name: 'Add parameter'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(api.bodies('POST /p1/app/1/parameter')).toEqual([{name: 'Region', type: {type: 'option', defaultValue: 'us', options: ['eu', 'us']}}]);
});

it('Settings > Apps > parameters: adds options with the Add button and drops the default of a removed option', async () => {
    const {user} = await open();
    await user.click(screen.getByRole('button', {name: 'Add parameter'}));
    const d = await dialog('Add a parameter');
    await user.click(within(d).getByRole('radio', {name: 'Choice of options'}));
    await user.type(within(d).getByLabelText('Option'), 'x');
    await user.click(within(d).getByRole('button', {name: 'Add'}));
    await user.click(within(d).getByRole('combobox', {name: /Default/}));
    await user.click(screen.getByRole('option', {name: 'x'}));
    await removeChip(user, d, 'x');
    expect(within(d).queryByRole('button', {name: 'x'})).toBeNull();
    expect(within(d).getByRole('combobox', {name: /Default/})).toHaveTextContent(/^\u200b?$/);
});

it('Settings > Apps > parameters: checks the form before saving', async () => {
    const {user} = await open();
    await user.click(screen.getByRole('button', {name: 'Add parameter'}));
    const d = await dialog('Add a parameter');
    await user.type(within(d).getByLabelText('Name'), 'mode');
    await user.click(within(d).getByRole('radio', {name: 'Choice of options'}));
    await user.click(within(d).getByRole('button', {name: 'Add parameter'}));
    expect(within(d).getByText('The app has a parameter with this name.')).toBeInTheDocument();
    expect(within(d).getByText('A choice needs at least one option.')).toBeInTheDocument();
});

it('Settings > Apps > parameters: edits a parameter', async () => {
    const updated = parameter({name: 'MarketId', type: {type: 'string', defaultValue: 'us'}});
    const {api, user} = await open({'PUT /p1/app/1/parameter/10': () => updated});
    await user.click(screen.getByRole('button', {name: 'Edit MarketId'}));
    const d = await dialog('Edit MarketId');
    const field = within(d).getByLabelText('Default');
    await user.clear(field);
    await user.type(field, 'us');
    await user.click(within(d).getByRole('button', {name: 'Save'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(api.bodies('PUT /p1/app/1/parameter/10')).toEqual([{name: 'MarketId', type: {type: 'string', defaultValue: 'us'}}]);
    expect(screen.getByText('us')).toBeInTheDocument();
});

it('Settings > Apps > parameters: shows the master\'s error when saving fails', async () => {
    const {user} = await open({'PUT /p1/app/1/parameter/10': () => {
        throw apiError('rejected');
    }});
    await user.click(screen.getByRole('button', {name: 'Edit MarketId'}));
    const d = await dialog('Edit MarketId');
    await user.click(within(d).getByRole('button', {name: 'Save'}));
    expect(await within(d).findByText('rejected')).toBeInTheDocument();
});

it('Settings > Apps > parameters: deletes a parameter after a confirmation', async () => {
    const {api, user} = await open({'DELETE /p1/app/1/parameter/10': () => undefined});
    await user.click(screen.getByRole('button', {name: 'Delete MarketId'}));
    const d = await dialog('Delete MarketId?');
    expect(within(d).getByText(/Runs that already used it keep their values/)).toBeInTheDocument();
    await user.click(within(d).getByRole('button', {name: 'Delete parameter'}));
    await waitFor(() => expect(screen.queryByText('MarketId')).toBeNull());
    expect(api.calls.map(c => c.key)).toContain('DELETE /p1/app/1/parameter/10');
});

it('Settings > Apps > parameters: keeps the parameter and shows the error when the delete fails', async () => {
    const {user} = await open({'DELETE /p1/app/1/parameter/10': () => {
        throw apiError('locked');
    }});
    await user.click(screen.getByRole('button', {name: 'Delete MarketId'}));
    const d = await dialog('Delete MarketId?');
    await user.click(within(d).getByRole('button', {name: 'Delete parameter'}));
    expect(await screen.findByRole('alert', {hidden: true})).toHaveTextContent('locked');
    expect(screen.getByText('MarketId')).toBeInTheDocument();
});
