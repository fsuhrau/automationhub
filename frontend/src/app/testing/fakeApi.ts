import {ApiClient} from '../../api/ApiClient';

type Answer = unknown | (() => unknown);

export interface FakeApi extends ApiClient {
    // the urls that were asked, in order
    calls: string[],
}

// A fake ApiClient for tests: answers get() by url (a value, or a function that
// returns one or throws); unknown urls reject like a missing route.
export const fakeApi = (answers: Record<string, Answer>): FakeApi => {
    const calls: string[] = [];
    const answer = async <T, >(url: string): Promise<T> => {
        calls.push(url);
        if (!(url in answers)) throw new Error(`no fake answer for ${url}`);
        const value = answers[url];
        return (typeof value === 'function' ? value() : value) as T;
    };
    return {calls, get: answer, post: answer, put: answer, delete: answer};
};
