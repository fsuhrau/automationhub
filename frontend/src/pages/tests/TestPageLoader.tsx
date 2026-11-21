import React from 'react';
import {useParams} from 'react-router-dom';
import EditTestPage from './EditTestPage';
import ShowTestPage from './ShowTestPage';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useLoader} from '../../hooks/useLoader';
import PageLoader from '../../components/PageLoader';
import {useTestService} from '../../services/test.service';

interface TestPageProps {
    edit: boolean
}

const TestPageLoader: React.FC<TestPageProps> = ({edit}) => {
    const {projectIdentifier} = useProjectContext();
    const {testId} = useParams();
    const {appId} = useApplicationContext();
    const service = useTestService();

    const {data: test, error, retry} = useLoader(
        () => appId !== null && testId !== undefined ? service.get({projectId: projectIdentifier, appId}, testId) : null,
        [service, projectIdentifier, appId, testId]);

    if (!test || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    return edit ? <EditTestPage test={test}/> : <ShowTestPage test={test}/>;
};

export default TestPageLoader;
