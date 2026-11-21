import {describe, expect, it} from 'vitest';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';
import {aFunction, aTest} from './definitions/harness/fixtures';
import {
    addFunctions, canSubmit, formOf, functionKey, groupByClass, isDirty, matchesFunction, mergeCategories, newForm, noChoicesText,
    parsePastedTests, saveDisabled, saveHint, SaveState, shortClass, shortMethod, splitCategories, toggled, toRequest, validate,
} from './testEditModel';

const state = (over: Partial<SaveState> = {}): SaveState => ({creating: false, dirty: true, tried: false, valid: true, ...over});

describe('splitCategories', () => {
    it.each([
        ['a, b ,,a', ['a', 'b']],
        ['', []],
        [undefined, []],
        [null, []],
    ])('splits %j', (text, expected) => {
        expect(splitCategories(text)).toEqual(expected);
    });
});

describe('formOf and newForm', () => {
    it('reads the form of a stored test', () => {
        const test = aTest({}, {
            allDevices: false, recordSession: true, executionType: TestExecutionType.Simultaneously,
            devices: [{deviceId: 4, testConfigId: 1, createdAt: new Date(), updatedAt: new Date(), deletedAt: new Date()}],
        });
        test.testConfig.unity = {
            ...test.testConfig.unity!, testCategoryType: UnityTestCategory.RunAllOfCategory, categories: 'a,b',
            testFunctions: [{...aFunction('M'), id: 9, testConfigUnityId: 1}],
        };
        expect(formOf(test)).toEqual({
            name: 'Smoke', executionType: TestExecutionType.Simultaneously, allDevices: false, devices: [4],
            mode: UnityTestCategory.RunAllOfCategory, categories: ['a', 'b'], functions: [aFunction('M')], recordSession: true,
        });
    });

    it('falls back to all Unity tests without a unity config', () => {
        const test = aTest();
        test.testConfig.unity = null;
        test.testConfig.devices = null as never;
        expect(formOf(test)).toMatchObject({mode: UnityTestCategory.RunAllTests, devices: [], functions: [], recordSession: false});
    });

    it('starts a new test with all tests on all devices', () => {
        expect(newForm()).toMatchObject({name: '', allDevices: true, mode: UnityTestCategory.RunAllTests, executionType: TestExecutionType.Concurrent});
    });
});

describe('functions', () => {
    it('addFunctions keeps the order and skips the picked ones', () => {
        const a = aFunction('A');
        const b = aFunction('B');
        expect(addFunctions([a], [a, b, b])).toEqual([a, b]);
    });

    it('functionKey joins the three names', () => {
        expect(functionKey(aFunction('M'))).toBe('Game|Ns.Fight|M');
    });

    it('shortClass and shortMethod drop the namespace and the return type', () => {
        expect(shortClass('Ns.Sub.Fight')).toBe('Fight');
        expect(shortMethod(' System.Void Run ')).toBe('Run');
    });

    it('groupByClass groups in the order the classes appear', () => {
        const groups = groupByClass([aFunction('A', 'X'), aFunction('B', 'Y'), aFunction('C', 'X')]);
        expect(groups.map(([cls, fns]) => [cls, fns.map(f => f.method)])).toEqual([['X', ['A', 'C']], ['Y', ['B']]]);
    });

    it.each([
        ['', true],
        ['fight', true],
        ['GAME', true],
        ['nope', false],
    ])('matchesFunction %j is %s', (text, expected) => {
        expect(matchesFunction(aFunction('Run'), text)).toBe(expected);
    });

    it('matchesFunction copes with a missing assembly', () => {
        expect(matchesFunction({...aFunction('Run'), assembly: undefined as never}, 'zzz')).toBe(false);
    });
});

describe('parsePastedTests', () => {
    it('reads a list and fills the assembly', () => {
        const result = parsePastedTests('[{"class":"C","method":"M"},{"nope":1},null]');
        expect(result).toEqual({functions: [{assembly: '', class: 'C', method: 'M'}]});
    });

    it.each([
        ['not json', "That isn't JSON"],
        ['{}', 'Expected a list of tests.'],
        ['[{"a":1}]', 'The list has no tests'],
    ])('rejects %j', (text, message) => {
        expect(parsePastedTests(text)).toEqual({error: expect.stringContaining(message)});
    });
});

describe('validate', () => {
    it('accepts a new form once it has a name', () => {
        expect(validate({...newForm(), name: 'x'}, true)).toEqual({});
    });

    it('lists what is missing', () => {
        const picked = {...newForm(), name: ' ', allDevices: false, mode: UnityTestCategory.RunSelectedTestsOnly};
        expect(Object.keys(validate(picked, true))).toEqual(['name', 'devices', 'tests']);
        expect(Object.keys(validate({...picked, mode: UnityTestCategory.RunAllOfCategory}, true))).toEqual(['name', 'devices', 'categories']);
    });

    it('ignores the Unity rules for other test types', () => {
        expect(validate({...newForm(), name: 'x', mode: UnityTestCategory.RunSelectedTestsOnly}, false)).toEqual({});
    });
});

describe('isDirty and toRequest', () => {
    it('ignores the order and the whitespace', () => {
        const a = {...newForm(), name: 'x', allDevices: false, devices: [1, 2], categories: ['a', 'b']};
        expect(isDirty(a, {...a, name: ' x ', devices: [2, 1], categories: ['b', 'a']})).toBe(false);
        expect(isDirty(a, {...a, recordSession: true})).toBe(true);
    });

    it('sends no devices when all are used and trims the name', () => {
        const form = {...newForm(), name: ' x ', devices: [1], categories: ['a', 'b'], mode: UnityTestCategory.RunAllOfCategory};
        expect(toRequest(form)).toMatchObject({name: 'x', devices: [], categories: 'a,b', unityTestCategoryType: UnityTestCategory.RunAllOfCategory});
        expect(toRequest({...form, allDevices: false}).devices).toEqual([1]);
    });
});

describe('small helpers', () => {
    it('mergeCategories adds without doubles', () => {
        expect(mergeCategories(['a'], 'b, a,c')).toEqual(['a', 'b', 'c']);
    });

    it('toggled adds and removes without changing the original', () => {
        const base = new Set(['a']);
        expect([...toggled(base, 'b', true)]).toEqual(['a', 'b']);
        expect([...toggled(base, 'a', false)]).toEqual([]);
        expect([...base]).toEqual(['a']);
    });

    it.each([
        [0, '', 'Unity reported no tests.'],
        [3, 'x', 'No test matches.'],
        [3, '', 'Every test is picked.'],
    ])('noChoicesText(%s, %j)', (available, search, expected) => {
        expect(noChoicesText(available, search)).toBe(expected);
    });
});

describe('saving rules', () => {
    it('canSubmit needs a valid form, and for a known test a change', () => {
        expect(canSubmit(state())).toBe(true);
        expect(canSubmit(state({valid: false}))).toBe(false);
        expect(canSubmit(state({dirty: false}))).toBe(false);
        expect(canSubmit(state({dirty: false, creating: true}))).toBe(true);
    });

    it('saveDisabled and saveHint say why saving is not possible', () => {
        expect(saveDisabled(state({dirty: false}))).toBe(true);
        expect(saveHint(state({dirty: false}))).toBe('Nothing to save yet');
        expect(saveDisabled(state({tried: true, valid: false}))).toBe(true);
        expect(saveHint(state({tried: true, valid: false}))).toBe('Fix the marked fields first');
        expect(saveDisabled(state())).toBe(false);
        expect(saveHint(state())).toBe('Save (⌘S / Ctrl+S)');
        expect(saveHint(state({creating: true}))).toBe('Create (⌘S / Ctrl+S)');
    });
});
