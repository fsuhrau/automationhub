import userEvent from '@testing-library/user-event';
import {FakeApi} from './fakeApi';

export type User = ReturnType<typeof userEvent.setup>;

// what a test gets after opening a page: the fake master and the person using it
export interface Opened {
    api: FakeApi,
    user: User,
}
