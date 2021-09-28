import { FC, useEffect, useState } from 'react';
import { getLastRun } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestProtocolData from '../../types/test.protocol';
import TestProtocolContent from '../../components/testprotocol-content.component';
import ITestRunData from '../../types/test.run';
import { Typography } from '@material-ui/core';

const TestProtocol: FC = () => {
    const { testId } = useParams<number>();
    const { protocolId } = useParams<number>();

    const [protocol, setProtocol] = useState<ITestProtocolData>();
    const [run, setRun] = useState<ITestRunData>();

    useEffect(() => {
        getLastRun(testId).then(response => {
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
    }, [testId, protocolId]);

    return (
        <div>
            { protocol && run
                ? <TestProtocolContent run={ run } protocol={ protocol }/>
                : <Typography variant={ 'h1' }>Loading</Typography>
            }
        </div>
    );
};

export default TestProtocol;
