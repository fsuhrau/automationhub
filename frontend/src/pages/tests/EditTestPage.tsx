import React, {useState} from 'react';
import {Stack} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import ITestData from '../../types/test';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {saveDisabled, saveHint} from './testEditModel';
import {DiscardDialog} from './definitions/parts/DiscardDialog';
import {EditHeader} from './definitions/parts/EditHeader';
import {ErrorSummary} from './definitions/parts/ErrorSummary';
import {FormShell} from './definitions/parts/FormShell';
import {NameSection} from './definitions/parts/NameSection';
import {WhatItRunsSection} from './definitions/parts/WhatItRunsSection';
import {WhereItRunsSection} from './definitions/parts/WhereItRunsSection';
import {useAppRoutes} from './definitions/useAppRoutes';
import {useSaveTest} from './definitions/useSaveTest';
import {useTestDevices} from './definitions/useTestDevices';
import {useTestForm} from './definitions/useTestForm';
import {useUnsavedWarning} from './definitions/useUnsavedWarning';

// EditTestPage creates or changes a test: its name, what it runs and where.
// Without a test it creates a new Unity test.
const EditTestPage: React.FC<{ test?: ITestData }> = ({test}) => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const navigate = useNavigate();
    const appRoutes = useAppRoutes();
    const app = state.apps?.find(a => a.id === appId);
    // where Cancel leads: the test, or the list for a new one
    const backUrl = test ? appRoutes.test(test.id).show() : appRoutes.tests();

    const formState = useTestForm(test);
    const {form, set, errors, dirty, tried, unity, saveState} = formState;
    const devices = useTestDevices(projectIdentifier, app?.platform);
    const {submit, saving} = useSaveTest(test, formState);
    const [leaving, setLeaving] = useState(false);
    useUnsavedWarning(dirty);
    const shown = tried ? errors : {};

    return (
        <FormShell onSubmit={submit}>
            <EditHeader test={test} appName={app?.name} dirty={dirty} saving={saving} saveHint={saveHint(saveState)}
                saveDisabled={saveDisabled(saveState)} onCancel={() => dirty ? setLeaving(true) : navigate(backUrl)}/>
            <ErrorSummary errors={errors} tried={tried}/>
            <Stack spacing={2}>
                <NameSection name={form.name} onChange={name => set({name})} creating={!test} error={shown.name}/>
                {unity && <WhatItRunsSection form={form} set={set} errors={shown}/>}
                <WhereItRunsSection form={form} set={set} errors={shown} devices={devices} appName={app?.name}/>
            </Stack>
            <DiscardDialog open={leaving} testName={test?.name} onKeep={() => setLeaving(false)} onDiscard={() => navigate(backUrl)}/>
        </FormShell>
    );
};

export default EditTestPage;
