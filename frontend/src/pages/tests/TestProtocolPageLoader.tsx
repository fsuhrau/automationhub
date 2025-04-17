import React, {useEffect, useState} from 'react';
import {getTestProtocol} from '../../services/test.run.service';
import {useParams} from 'react-router-dom';
import ITestProtocolData from '../../types/test.protocol';
import TestProtocolPage from './TestProtocolPage';
import ITestRunData from '../../types/test.run';
import {Backdrop, CircularProgress} from '@mui/material';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {useError} from "../../ErrorProvider";

const TestProtocolPageLoader: React.FC = () => {

    const {projectIdentifier} = useProjectContext();

    const {testId, runId, protocolId} = useParams();
    const {appId} = useApplicationContext()
    const {setError} = useError()

    const [state, setState] = useState<{
        loading: boolean,
        run: ITestRunData | null,
        protocol: ITestProtocolData | null
    }>({loading: true, run: null, protocol: null})

    useEffect(() => {
        if (protocolId === undefined) {
            return;
        }

        getTestProtocol(projectIdentifier, appId as number, testId as string, runId as string, protocolId as string).then(response => {
            const protocol = response.data.Protocols.find(p => p.ID === +protocolId)
            setState(prevState => ({
                ...prevState,
                loading: false,
                run: response.data,
                protocol: protocol === undefined ? null : protocol,
            }));
        }).catch(ex => {
            setError(ex);
        });
    }, [testId, runId, protocolId, appId]);

    if (state.loading) return (<Backdrop sx={{zIndex: 1}} open={true}>
        <CircularProgress color="inherit"/>
        </Backdrop>)

    return (<TestProtocolPage run={state.run as ITestRunData} protocol={state.protocol as ITestProtocolData}/>);
};

export default TestProtocolPageLoader;
