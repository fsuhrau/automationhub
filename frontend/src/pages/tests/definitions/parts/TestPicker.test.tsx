import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, describe, expect, it} from 'vitest';
import {failure, FakeApi} from '../testing/fakeApi';
import {aFunction} from '../testing/fixtures';
import {CREATE, openEditor, UNITY} from '../testing/editSupport';

const FUNCTIONS = [aFunction('System.Void Hit'), aFunction('System.Void Block'), aFunction('System.Void Walk', 'Ns.Move')];
let api: FakeApi;

const open = async (unity: unknown = FUNCTIONS): Promise<void> => {
    api = openEditor(null, {[UNITY]: unity as never}).api;
    await userEvent.click(await screen.findByRole('radio', {name: /Picked tests/}));
};
const search = (text: string): Promise<void> => userEvent.type(screen.getByRole('textbox', {name: 'Find a test to add'}), text);
const silent = (): never => {
    throw failure(500, 'no unity');
};
const paste = (text: string): Promise<void> => userEvent.click(screen.getByRole('textbox', {name: 'Pasted test list'})).then(() => userEvent.paste(text));
const picked = (): HTMLElement => screen.getByText(/^Picked tests \(/);

afterEach(() => api.restore());

describe('TestPicker with the Unity editor', () => {
    it('adds all tests Unity reported', async () => {
        await open();
        await userEvent.click(await screen.findByRole('button', {name: 'Add all 3'}));
        expect(picked()).toHaveTextContent('Picked tests (3)');
        expect(screen.getByText('Every test is picked.')).toBeInTheDocument();
    });

    it('adds the checked tests only', async () => {
        await open();
        const group = await screen.findByRole('group', {name: 'Tests to add'});
        expect(screen.getByRole('button', {name: 'Add picked'})).toBeDisabled();
        await userEvent.click(within(group).getByRole('checkbox', {name: 'Hit'}));
        await userEvent.click(within(group).getByRole('checkbox', {name: 'Walk'}));
        await userEvent.click(within(group).getByRole('checkbox', {name: 'Walk'}));
        await userEvent.click(screen.getByRole('button', {name: 'Add 1 picked'}));
        expect(picked()).toHaveTextContent('Picked tests (1)');
        expect(screen.queryByRole('checkbox', {name: 'Hit'})).not.toBeInTheDocument();
    });

    it('finds tests by name and says when none matches', async () => {
        await open();
        await screen.findByRole('group', {name: 'Tests to add'});
        await search('blo');
        expect(screen.getByRole('button', {name: 'Add all 1 found'})).toBeInTheDocument();
        await userEvent.clear(screen.getByRole('textbox', {name: 'Find a test to add'}));
        await search('zzz');
        expect(screen.getByText('No test matches.')).toBeInTheDocument();
    });

});

describe('TestPicker with the Unity editor (more)', () => {
    it('removes a picked test again and groups the picked by class', async () => {
        await open();
        await userEvent.click(await screen.findByRole('button', {name: 'Add all 3'}));
        expect(screen.getByText('Fight')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Remove Hit'}));
        expect(picked()).toHaveTextContent('Picked tests (2)');
        expect(screen.getByRole('checkbox', {name: 'Hit'})).toBeInTheDocument();
    });

    it('says when Unity has no tests', async () => {
        await open([]);
        expect(await screen.findByText('Unity reported no tests.')).toBeInTheDocument();
    });

    it('asks Unity again with the refresh button', async () => {
        let list = [FUNCTIONS[0]];
        await open(() => list);
        await screen.findByRole('checkbox', {name: 'Hit'});
        list = FUNCTIONS;
        await userEvent.click(screen.getByRole('button', {name: 'Load the tests from Unity again'}));
        expect(await screen.findByRole('checkbox', {name: 'Walk'})).toBeInTheDocument();
    });

    it('shows that it is asking Unity', async () => {
        await open(() => new Promise(() => undefined));
        expect(await screen.findByText(/Asking the Unity editor/)).toBeInTheDocument();
    });

});

describe('TestPicker creating', () => {
    it('needs at least one test to create', async () => {
        await open();
        await userEvent.type(screen.getByRole('textbox', {name: 'Name'}), 'Fights');
        await userEvent.click(screen.getByRole('button', {name: 'Create test'}));
        expect(await screen.findAllByText('Pick at least one test.')).toHaveLength(2);
    });

    it('creates the test with the picked tests', async () => {
        await open();
        await userEvent.type(screen.getByRole('textbox', {name: 'Name'}), 'Fights');
        await userEvent.click(await screen.findByRole('button', {name: 'Add all 3'}));
        await userEvent.click(screen.getByRole('button', {name: 'Create test'}));
        await waitFor(() => expect(api.calledWith(CREATE)).toHaveLength(1));
        expect((api.calledWith(CREATE)[0].data as {testFunctions: unknown[]}).testFunctions).toHaveLength(3);
    });
});

describe('TestPicker without the Unity editor', () => {
    it('asks to open Unity or paste the list', async () => {
        await open(silent);
        expect(await screen.findByText(/The Unity editor on this computer didn't answer/)).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Use this list'})).toBeDisabled();
    });

    it.each([
        ['text that is not JSON', 'nope', "That isn't JSON"],
        ['JSON that is not a list', '{}', 'Expected a list of tests.'],
        ['a list without tests', '[1]', 'The list has no tests'],
    ])('rejects %s', async (_, text, message) => {
        await open(silent);
        await screen.findByRole('textbox', {name: 'Pasted test list'});
        await paste(text);
        await userEvent.click(screen.getByRole('button', {name: 'Use this list'}));
        expect(await screen.findByText(new RegExp(message))).toBeInTheDocument();
    });

});

describe('TestPicker with a pasted list', () => {
    it('uses a pasted list to add tests from', async () => {
        await open(silent);
        await screen.findByRole('textbox', {name: 'Pasted test list'});
        await paste('[{"assembly":"A","class":"C","method":"Jump"}]');
        await userEvent.click(screen.getByRole('button', {name: 'Use this list'}));
        expect(await screen.findByRole('checkbox', {name: 'Jump'})).toBeInTheDocument();
        expect(screen.queryByText(/didn't answer/)).not.toBeInTheDocument();
    });

    it('clears the paste error after a good list', async () => {
        await open(silent);
        const box = await screen.findByRole('textbox', {name: 'Pasted test list'});
        await paste('nope');
        await userEvent.click(screen.getByRole('button', {name: 'Use this list'}));
        await screen.findByText(/That isn't JSON/);
        await userEvent.clear(box);
        await paste('[{"class":"C","method":"Jump"}]');
        await userEvent.click(screen.getByRole('button', {name: 'Use this list'}));
        expect(await screen.findByRole('checkbox', {name: 'Jump'})).toBeInTheDocument();
    });
});
