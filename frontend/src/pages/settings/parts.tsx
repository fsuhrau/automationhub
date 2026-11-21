// The building blocks of the settings tabs moved to components/common; this
// re-exports them so the tabs keep their imports until F13 switches them.
export {Code, ConfirmDialog, CopyButton, Section} from '../../components/common';
export {errorMessage as messageOf} from '../../lib/errorMessage';
