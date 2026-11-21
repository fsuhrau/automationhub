import React from 'react';
import {Alert, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, Radio, RadioGroup, Stack, TextField} from '@mui/material';
import {useAsyncAction} from '../../../hooks/useAsyncAction';
import {AppParameter} from '../../../types/app';
import {ParameterForm} from '../settingsModel';
import {useParameterForm} from '../useParameterForm';
import {DefaultField, OptionsEditor} from './ParameterFields';

interface ParameterFormProps {
    parameter: AppParameter | null,
    siblings: AppParameter[],
    onSave: (parameter: AppParameter | null, form: ParameterForm) => Promise<boolean>,
    onClose: () => void,
}

// the form's state lives here, so every opening of the dialog starts fresh
const ParameterFormBody: React.FC<ParameterFormProps> = ({parameter, siblings, onSave, onClose}) => {
    const state = useParameterForm(parameter, siblings);
    const {form, set, errors} = state;
    const action = useAsyncAction(onSave);
    const submit = async (): Promise<void> => {
        if (state.validate() && await action.run(parameter, form)) onClose();
    };

    return <Box component="form" noValidate={true} onSubmit={e => {
        e.preventDefault();
        void submit();
    }}>
        <DialogTitle id="param-title">{parameter ? `Edit ${parameter.name}` : 'Add a parameter'}</DialogTitle>
        <DialogContent>
            <Stack spacing={2} sx={{pt: 1}}>
                <TextField size="small" autoFocus={true} label="Name" placeholder="e.g. MarketId" value={form.name}
                    onChange={e => set({name: e.target.value})} error={!!errors.name}
                    helperText={errors.name ?? 'Passed to the app as NAME=value when a test runs.'}/>
                <RadioGroup row={true} value={form.kind} onChange={e => set({kind: e.target.value as ParameterForm['kind']})}>
                    <FormControlLabel value="string" control={<Radio size="small"/>} label="Text"/>
                    <FormControlLabel value="option" control={<Radio size="small"/>} label="Choice of options"/>
                </RadioGroup>
                {form.kind === 'option' && <OptionsEditor {...state}/>}
                <DefaultField form={form} set={set} options={state.options} errors={errors}/>
                {action.error && <Alert severity="error">{action.error}</Alert>}
            </Stack>
        </DialogContent>
        <DialogActions>
            <Button onClick={onClose}>Cancel</Button>
            <Button type="submit" variant="contained" disabled={action.busy}>{parameter ? 'Save' : 'Add parameter'}</Button>
        </DialogActions>
    </Box>;
};

export interface ParameterDialogProps extends ParameterFormProps {
    open: boolean,
}

const ParameterDialog: React.FC<ParameterDialogProps> = ({open, ...form}) =>
    <Dialog open={open} onClose={form.onClose} fullWidth={true} maxWidth="sm" aria-labelledby="param-title">
        <ParameterFormBody {...form}/>
    </Dialog>;

export default ParameterDialog;
