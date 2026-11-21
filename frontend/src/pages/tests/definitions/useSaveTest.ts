import {useCallback, useEffect} from 'react';
import {useNavigate} from 'react-router-dom';
import {useError} from '../../../ErrorProvider';
import {useApplicationContext} from '../../../hooks/ApplicationProvider';
import {useAsyncAction} from '../../../hooks/useAsyncAction';
import {useProjectContext} from '../../../hooks/ProjectProvider';
import ITestData from '../../../types/test';
import {TestType} from '../../../types/test.type.enum';
import {canSubmit, TestForm, toRequest} from '../testEditModel';
import {TestFormState} from './useTestForm';
import {useAppRoutes} from './useAppRoutes';
import {useTestService} from '../../../services/test.service';

export interface SaveTest {
    // marks the form as tried and saves it when it may be saved
    submit: () => void,
    saving: boolean,
}

// useSaveTest creates a new Unity test or updates the test, then opens it.
// A failure is shown as the page error.
export const useSaveTest = (test: ITestData | undefined, state: TestFormState): SaveTest => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const appRoutes = useAppRoutes();
    const service = useTestService();
    const navigate = useNavigate();
    const {setError} = useError();

    const send = useCallback(async (form: TestForm): Promise<void> => {
        if (appId === null) return;
        const id = test
            ? (await service.update({projectId: projectIdentifier, appId}, test.id, toRequest(form)), test.id)
            : (await service.create({projectId: projectIdentifier, appId}, {type: TestType.Unity, ...toRequest(form)})).id;
        navigate(appRoutes.test(id).show());
    }, [service, projectIdentifier, appId, test, appRoutes, navigate]);
    const {run, busy, error} = useAsyncAction(send);
    const {form, saveState, markTried} = state;
    useEffect(() => {
        if (error) setError(error);
    }, [error, setError]);
    return {
        saving: busy,
        submit: () => {
            markTried();
            if (canSubmit(saveState)) void run(form);
        },
    };
};
