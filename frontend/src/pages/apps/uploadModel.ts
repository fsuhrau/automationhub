// What a bundle upload shows: the text, the alert color and the transitions
// between the states of an upload.
import {errorMessage} from '../../lib/errorMessage';
import {UploadState} from '../../types/bundle.upload';
import {Accept} from './bundlesModel';

type Phase = UploadState['phase'];
type Of<P extends Phase> = Extract<UploadState, { phase: P }>;
export type Severity = 'info' | 'success' | 'warning' | 'error';

const TEXT: { [P in Phase]: (s: Of<P>) => string } = {
    sending: s => `Uploading ${Math.round(s.sent * 100)} %`,
    reading: () => 'Reading the bundle on the master…',
    done: s => `Uploaded as version ${s.bundle.version || s.bundle.name} (#${s.bundle.id})`,
    exists: () => 'This exact file was uploaded before; the list already has it.',
    failed: s => s.message,
    cancelled: () => 'Upload cancelled.',
};

const SEVERITY: Record<Phase, Severity> = {
    sending: 'info', reading: 'info', done: 'success', exists: 'warning', failed: 'error', cancelled: 'warning',
};

export const uploadText = (state: UploadState): string => (TEXT[state.phase] as (s: UploadState) => string)(state);

export const uploadSeverity = (state: UploadState): Severity => SEVERITY[state.phase];

// isActive: the file is still on its way or being read
export const isActive = (state: UploadState): boolean => state.phase === 'sending' || state.phase === 'reading';

export const CANCELLED: UploadState = {phase: 'cancelled'};

// sendingState: all bytes sent means the master reads the file now
export const sendingState = (sent: number): UploadState => sent < 1 ? {phase: 'sending', sent} : {phase: 'reading'};

const statusOf = (e: unknown): number | undefined =>
    (e as { response?: { status?: number } } | null)?.response?.status;

// failureState says how an upload ended that threw; an identical file is a 409
export const failureState = (e: unknown, aborted: boolean): UploadState => {
    if (aborted) return CANCELLED;
    if (statusOf(e) === 409) return {phase: 'exists'};
    return {phase: 'failed', message: errorMessage(e)};
};

// rejectedState is for a file the app can't take, which is never sent
export const rejectedState = (appName: string | undefined, accept: Accept): UploadState =>
    ({phase: 'failed', message: `${appName ?? 'This app'} takes ${accept.label} files.`});
