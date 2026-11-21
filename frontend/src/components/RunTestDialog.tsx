import React from 'react';
import {IAppData} from '../types/app';
import {RunTestForm} from '../pages/tests/definitions/runDialog/RunTestForm';

interface RunTestDialogProps {
    projectIdentifier: string,
    app: IAppData,
    test: { id: number, name: string } | null,
    onClose: () => void,
}

// RunTestDialog starts a test: it picks the bundle to install (the newest by
// default, or a new upload), takes the app's parameters and, for web apps,
// the start URL, then opens the new run. It is open while a test is given.
const RunTestDialog: React.FC<RunTestDialogProps> = ({test, ...rest}) =>
    test ? <RunTestForm test={test} {...rest}/> : null;

export default RunTestDialog;
