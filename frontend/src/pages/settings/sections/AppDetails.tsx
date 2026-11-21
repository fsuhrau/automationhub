import React from 'react';
import {Stack} from '@mui/material';
import {useProjectContext} from '../../../hooks/ProjectProvider';
import {IAppData} from '../../../types/app';
import AppFacts from './AppFacts';
import AppNameForm, {AppNameFormProps} from './AppNameForm';
import ParametersBlock, {ParametersBlockProps} from './ParametersBlock';

export interface AppDetailsProps extends Omit<ParametersBlockProps, 'parameters'> {
    app: IAppData,
    onRename: AppNameFormProps['onSave'],
}

// AppDetails is the selected app: its name, facts and parameters
const AppDetails: React.FC<AppDetailsProps> = ({app, onRename, ...parameters}) => {
    const {projectIdentifier} = useProjectContext();
    return <Stack spacing={2.5} sx={{minWidth: 0}}>
        <AppNameForm key={app.id} name={app.name} onSave={onRename}/>
        <AppFacts app={app} projectIdentifier={projectIdentifier}/>
        <ParametersBlock parameters={app.parameter ?? []} {...parameters}/>
    </Stack>;
};

export default AppDetails;
