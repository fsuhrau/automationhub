import {useReportedAction} from './useReportedAction';

export interface BundleCommand {
    busy: boolean,
    // resolves true when op worked; a failure shows in the error banner
    run: (op: () => Promise<unknown>) => Promise<boolean>,
}

// useBundleCommand runs one change of a bundle at a time for a button
export const useBundleCommand = (): BundleCommand => {
    const {run, busy} = useReportedAction(async (op: () => Promise<unknown>): Promise<boolean> => {
        await op();
        return true;
    });
    return {busy, run: async op => (await run(op)) === true};
};
