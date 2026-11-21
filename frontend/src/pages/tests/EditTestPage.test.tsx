import {fireEvent, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, describe, expect, it} from 'vitest';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {TestType} from '../../types/test.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';
import {failure, FakeApi} from './definitions/testing/fakeApi';
import {aDevice, aFunction, aTest} from './definitions/testing/fixtures';
import {DEVICES, location, openEditor, UPDATE} from './definitions/testing/editSupport';

let api: FakeApi;
const edit = async (test = aTest(), routes = {}): Promise<void> => {
    api = openEditor(test, routes).api;
    await screen.findByRole('heading', {name: `Edit ${test.name}`});
};
const nameField = (): HTMLElement => screen.getByRole('textbox', {name: 'Name'});
const save = (): Promise<void> => userEvent.click(screen.getByRole('button', {name: 'Save'}));
const body = (): unknown => api.calledWith(UPDATE)[0].data;
const picked = (): Promise<void> => userEvent.click(screen.getByRole('radio', {name: /Picked devices/}));
const devices = [aDevice({id: 5, alias: 'Lab phone'}), aDevice({id: 6, name: 'Pixel'})];
const device = (deviceId: number): NonNullable<ReturnType<typeof aTest>['testConfig']['devices']>[number] =>
    ({deviceId, testConfigId: 1, createdAt: new Date(), updatedAt: new Date(), deletedAt: new Date()});

afterEach(() => api.restore());

describe('EditTestPage', () => {
    it('shows the test with nothing to save yet', async () => {
        await edit();
        expect(screen.getByText('Unity test · no changes yet')).toBeInTheDocument();
        expect(nameField()).toHaveValue('Smoke');
        expect(screen.getByRole('button', {name: 'Save'})).toBeDisabled();
    });

    it('saves the changed name and opens the test', async () => {
        await edit();
        await userEvent.type(nameField(), ' 2');
        expect(screen.getByText('Unity test · unsaved changes')).toBeInTheDocument();
        await save();
        await waitFor(() => expect(location()).toBe('/project/proj/app:1/test/7'));
        expect(body()).toEqual({
            name: 'Smoke 2', executionType: TestExecutionType.Concurrent, allDevices: true, devices: [], unityTestCategoryType: UnityTestCategory.RunAllTests,
            categories: '', testFunctions: [], recordSession: false,
        });
    });

    it('saves with Ctrl+S', async () => {
        await edit();
        await userEvent.type(nameField(), 'x');
        fireEvent.keyDown(nameField(), {key: 's', ctrlKey: true});
        await waitFor(() => expect(api.calledWith(UPDATE)).toHaveLength(1));
        fireEvent.keyDown(nameField(), {key: 'a', ctrlKey: true});
        expect(api.calledWith(UPDATE)).toHaveLength(1);
    });

    it('does not save without a change', async () => {
        await edit();
        fireEvent.submit(nameField().closest('form')!);
        expect(api.calledWith(UPDATE)).toHaveLength(0);
    });

});

describe('EditTestPage validation and errors', () => {
    it('marks an empty name and does not save', async () => {
        await edit();
        await userEvent.clear(nameField());
        await save();
        expect(await screen.findAllByText('The test needs a name.')).toHaveLength(2);
        expect(screen.getByRole('button', {name: 'Save'})).toBeDisabled();
        expect(api.calledWith(UPDATE)).toHaveLength(0);
    });

    it('shows the save error and lets the user save again', async () => {
        await edit(aTest(), {[UPDATE]: () => {
            throw failure(500, 'disk full');
        }});
        await userEvent.type(nameField(), 'x');
        await save();
        expect(await screen.findByRole('alert')).toHaveTextContent('disk full');
        expect(screen.getByRole('button', {name: 'Save'})).toBeEnabled();
    });

    it('changes how tests spread and records the session', async () => {
        await edit();
        await userEvent.click(screen.getByRole('radio', {name: /Every test on every device/}));
        await userEvent.click(screen.getByRole('switch', {name: /Record the session/}));
        await save();
        await waitFor(() => expect(body()).toMatchObject({executionType: TestExecutionType.Simultaneously, recordSession: true}));
    });

    it('hides the Unity settings for other test types', async () => {
        await edit(aTest({}, {type: TestType.Cocos, unity: null}));
        expect(screen.getByText('Cocos test · no changes yet')).toBeInTheDocument();
        expect(screen.queryByText('What it runs')).not.toBeInTheDocument();
    });
});

describe('EditTestPage categories', () => {
    const categories = async (): Promise<void> => {
        await edit();
        await userEvent.click(screen.getByRole('radio', {name: /Tests of categories/}));
    };

    it('adds categories with Enter and commas and removes one', async () => {
        await categories();
        const field = screen.getByRole('textbox', {name: 'Category'});
        await userEvent.type(field, 'fast, slow{Enter}');
        await userEvent.type(field, 'fast');
        await userEvent.click(screen.getByRole('button', {name: 'Add'}));
        await userEvent.click(screen.getByText('slow').closest('.MuiChip-root')!.querySelector('svg')!);
        await save();
        await waitFor(() => expect(body()).toMatchObject({categories: 'fast', unityTestCategoryType: UnityTestCategory.RunAllOfCategory}));
    });

    it('needs a category', async () => {
        await categories();
        await save();
        expect(await screen.findAllByText('Name at least one category.')).toHaveLength(2);
    });
});

describe('EditTestPage devices', () => {
    it('picks devices from the list and saves them', async () => {
        await edit(aTest(), {[DEVICES]: devices});
        await picked();
        await userEvent.click(await screen.findByRole('checkbox', {name: /Lab phone/}));
        await userEvent.click(screen.getByRole('checkbox', {name: /Pixel/}));
        await userEvent.click(screen.getByRole('checkbox', {name: /Pixel/}));
        await save();
        await waitFor(() => expect(body()).toMatchObject({allDevices: false, devices: [5]}));
    });

    it('needs a device when not running on all', async () => {
        await edit(aTest(), {[DEVICES]: devices});
        await picked();
        await save();
        expect(await screen.findAllByText('Pick at least one device, or run on all devices.')).toHaveLength(2);
    });

    it('says when the hub knows no devices', async () => {
        await edit();
        await picked();
        expect(await screen.findByText(/No Hero devices are known to the hub yet/)).toBeInTheDocument();
    });

});

describe('EditTestPage missing devices', () => {
    it('offers to remove picked devices the list does not have', async () => {
        await edit(aTest({}, {allDevices: false, devices: [device(5), device(99), device(98)]}), {[DEVICES]: devices});
        expect(await screen.findByText(/2 picked devices aren't in the list/)).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Remove'}));
        expect(screen.queryByText(/picked devices aren't/)).not.toBeInTheDocument();
        await save();
        await waitFor(() => expect(body()).toMatchObject({devices: [5]}));
    });

    it('says a single missing device in the singular', async () => {
        await edit(aTest({}, {allDevices: false, devices: [device(5), device(99)]}), {[DEVICES]: devices});
        expect(await screen.findByText(/One picked device isn't in the list/)).toBeInTheDocument();
    });

    it('shows the error when the devices cannot be loaded', async () => {
        await edit(aTest(), {[DEVICES]: () => {
            throw failure(500, 'no devices today');
        }});
        expect(await screen.findByRole('alert')).toHaveTextContent('no devices today');
    });
});

describe('EditTestPage leaving', () => {
    it('leaves a clean form without asking', async () => {
        await edit();
        await userEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        expect(location()).toBe('/project/proj/app:1/test/7');
    });

    it('asks before changes are thrown away', async () => {
        await edit();
        await userEvent.type(nameField(), 'x');
        await userEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        expect(await screen.findByText("The changes to Smoke aren't saved yet.")).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Keep editing'}));
        await waitFor(() => expect(screen.queryByText('Discard your changes?')).not.toBeInTheDocument());
        expect(location()).toBe('/project/proj/app:1/test/7/edit');
        await userEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        await userEvent.click(await screen.findByRole('button', {name: 'Discard'}));
        expect(location()).toBe('/project/proj/app:1/test/7');
    });

});

describe('EditTestPage unsaved changes', () => {
    it('warns the browser about unsaved changes only while there are some', async () => {
        await edit();
        const event = new Event('beforeunload', {cancelable: true});
        window.dispatchEvent(event);
        expect(event.defaultPrevented).toBe(false);
        await userEvent.type(nameField(), 'x');
        const dirty = new Event('beforeunload', {cancelable: true});
        window.dispatchEvent(dirty);
        expect(dirty.defaultPrevented).toBe(true);
    });

    it('keeps the picked tests of the stored test', async () => {
        const test = aTest();
        test.testConfig.unity = {
            ...test.testConfig.unity!, testCategoryType: UnityTestCategory.RunSelectedTestsOnly, testFunctions: [{...aFunction('Hit'), testConfigUnityId: 1}],
        };
        await edit(test);
        expect(screen.getByText('Picked tests (1)')).toBeInTheDocument();
    });
});

describe('TestPageLoader', () => {
    it('shows the load error and loads again on retry', async () => {
        let fail = true;
        api = openEditor(aTest(), {'GET /proj/app/1/test/7': () => {
            if (fail) throw failure(500, 'test is gone');
            return aTest();
        }}).api;
        expect(await screen.findByText('test is gone')).toBeInTheDocument();
        fail = false;
        await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
        expect(await screen.findByRole('heading', {name: 'Edit Smoke'})).toBeInTheDocument();
    });
});
