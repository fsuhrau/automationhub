import React, { useEffect, useState } from 'react';
import { getTestProtocol } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestProtocolData from '../../types/test.protocol';
import TestProtocolContent from '../../components/testprotocol-content.component';
import ITestRunData from '../../types/test.run';
import { Backdrop, CircularProgress } from '@mui/material';
import { useProjectContext } from "../../hooks/ProjectProvider";

const TestProtocolLoader: React.FC = () => {

    const {projectId} = useProjectContext();

    const { testId, runId, protocolId } = useParams();

    const [protocol, setProtocol] = useState<ITestProtocolData>();
    const [run, setRun] = useState<ITestRunData>();

    useEffect(() => {
        if (protocolId === undefined) {
            return;
        }

        getTestProtocol(projectId, appId, testId as string, runId as string, protocolId as string).then(response => {
            setRun(response.data);
            for (let i = 0; i < response.data.Protocols.length; ++i) {
                if (response.data.Protocols[ i ].ID == +protocolId) {
                    setProtocol(response.data.Protocols[ i ]);
                    break;
                }
            }
        }).catch(ex => {
            console.log(ex);
        });
    }, [testId, runId, protocolId]);

    return (
        <div>
            { protocol && run
                ? <TestProtocolContent run={ run } protocol={ protocol }/>
                : <Backdrop sx={ { zIndex: 1, color: '#fff' } } open={ true }><CircularProgress color="inherit"/></Backdrop>
            }
        </div>
    );
};

export default TestProtocolLoader;
