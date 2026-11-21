import {useCallback} from 'react';
import {useSearchParams} from 'react-router-dom';
import {errorMessage} from '../../lib/errorMessage';
import {IAppBinaryData, IAppBundleDetail} from '../../types/app';
import {AppRef} from '../../types/app.ref';
import {pendingPair, selectionOf, Selection, swapped, withPick} from './compareSelectionModel';
import {useBundleDetail} from './useBundleDetail';
import {useBundleList} from './useBundleList';

export interface BundleCompare {
    // newest first; null while loading
    bundles: IAppBinaryData[] | null,
    selection: Selection,
    // the two bundles once both are loaded
    from?: IAppBundleDetail,
    to?: IAppBundleDetail,
    error: string | null,
    // two bundles are there to compare, but their details are still loading
    loading: boolean,
    pick: (which: 'a' | 'b', id: number) => void,
    swap: () => void,
}

// the loader keeps the previous bundle while the next one loads
const settled = (detail: IAppBundleDetail | null, id: number | undefined): IAppBundleDetail | undefined =>
    detail && detail.id === id ? detail : undefined;

// useBundleCompare loads the bundles of the app and the two the url picks
export const useBundleCompare = (app: AppRef | null): BundleCompare => {
    const [params, setParams] = useSearchParams();
    const list = useBundleList(app);
    const selection = selectionOf(list.data, params.get('a'), params.get('b'));
    const from = useBundleDetail(app, selection.a ?? NaN);
    const to = useBundleDetail(app, selection.b ?? NaN);
    const failure = [list.error, from.error, to.error].find(Boolean);
    const pick = (which: 'a' | 'b', id: number): void => setParams(withPick(selection, which, id), {replace: true});
    const swap = useCallback((): void => setParams(swapped(selection), {replace: true}), [selection, setParams]);
    const [a, b] = [settled(from.data, selection.a), settled(to.data, selection.b)];
    const error = failure ? errorMessage(failure) : null;
    const loading = pendingPair(list.data, !!(a && b), !!error);
    return {bundles: list.data, selection, error, loading, from: a, to: b, pick, swap};
};
