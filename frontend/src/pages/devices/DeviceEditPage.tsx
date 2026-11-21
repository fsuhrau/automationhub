import React, {useState} from 'react';
import {Box, Stack, Typography} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import IDeviceData from '../../types/device';
import {ConfirmDialog} from '../../components/common';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {routes} from '../../lib/routes';
import {useDeviceEditForm} from './useDeviceEditForm';
import {useSaveDevice} from './useSaveDevice';
import {ActionError} from './parts/ActionError';
import {AliasSection} from './parts/AliasSection';
import {EditHeader} from './parts/EditHeader';
import {ParameterEditor} from './parts/ParameterEditor';
import {saveHandlers} from './parts/saveHandlers';

interface DeviceEditPageProps {
    device: IDeviceData
}

// full width like the other pages, so it lines up with the breadcrumbs;
// the form itself stays narrow enough to read
const formStyle = {width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2, '& > *': {maxWidth: 900}};

// DeviceEditPage changes what a person sets on a device: the alias and the
// custom parameters. The rest is reported by the device and isn't editable.
const DeviceEditPage: React.FC<DeviceEditPageProps> = ({device}) => {
    const navigate = useNavigate();
    const {projectIdentifier} = useProjectContext();
    const edit = useDeviceEditForm(device);
    const save = useSaveDevice(device.id);
    const [leaving, setLeaving] = useState(false);
    const valid = edit.errors.size === 0;
    const detailsUrl = routes.project(projectIdentifier).device(device.id).show();

    const submit = (): void => {
        if (valid && edit.dirty) void save.run(edit.form);
    };
    const cancel = (): void => {
        if (edit.dirty) setLeaving(true);
        else void navigate(detailsUrl);
    };

    return (
        <Box component="form" sx={formStyle} noValidate={true} {...saveHandlers(submit)}>
            <EditHeader device={device} dirty={edit.dirty} valid={valid} saving={save.busy} onCancel={cancel}/>
            <ActionError message={save.error}/>
            <Stack spacing={2}>
                <AliasSection name={device.name} alias={edit.form.alias} onChange={edit.setAlias}/>
                <ParameterEditor rows={edit.form.parameters} errors={edit.errors} onChange={edit.setRow}
                    onAdd={edit.addRow} onRemove={edit.removeRow}/>
                <Typography variant="body2" color="text.secondary">
                    The name, system, type and connection are reported by the device and its node, so they can't be edited here.
                </Typography>
            </Stack>
            <ConfirmDialog open={leaving} title="Discard your changes?" confirm="Discard" color="warning"
                onConfirm={() => void navigate(detailsUrl)} onClose={() => setLeaving(false)}>
                The alias and parameter changes on this page aren't saved yet.
            </ConfirmDialog>
        </Box>
    );
};

export default DeviceEditPage;
