import { FC, useEffect, useState } from 'react';
import { getTestProtocol } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestProtocolData from '../../types/test.protocol';
import TestProtocolContent from '../../components/testprotocol-content.component';
import ITestRunData from '../../types/test.run';
import { Backdrop, CircularProgress, Typography } from '@mui/material';

interface ParamTypes {
    testId: string
    runId: string
    protocolId: string
}

const TestProtocol: FC = () => {
    const { testId, runId, protocolId } = useParams<ParamTypes>();

    const [protocol, setProtocol] = useState<ITestProtocolData>();
    const [run, setRun] = useState<ITestRunData>();

    useEffect(() => {
        getTestProtocol(testId, runId, protocolId).then(response => {
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
                :  <Backdrop open={true} >
                    <CircularProgress color="inherit" />
                </Backdrop>
            }
        </div>
    );
};

export default TestProtocol;
