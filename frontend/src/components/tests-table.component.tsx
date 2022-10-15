import React, { ChangeEvent, useCallback, useEffect, useState } from 'react';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import ITestData from '../types/test';
import { executeTest, getAllTests } from '../services/test.service';
import {
    Button,
    ButtonGroup,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    TextField,
    Typography,
} from '@mui/material';
import { Edit, PlayArrow } from '@mui/icons-material';
import BinarySelection from './binary-selection.component';
import { useNavigate, useParams } from 'react-router-dom';
import { PlatformType } from '../types/platform.type.enum';
import { IAppBinaryData } from "../types/app";
import { ApplicationProps } from "../application/application.props";

interface TestTableProps extends ApplicationProps {
    appId: number | null
}

const TestsTable: React.FC<TestTableProps> = (props: TestTableProps) => {

    const { appId, appState } = props;

    let params = useParams();

    const navigate = useNavigate();

    const app = appState.project?.Apps.find(a => a.ID === appId);

    // dialog
    const [open, setOpen] = useState(false);
    const handleRunClickOpen = (): void => {
        setOpen(true);
    };
    const handleRunClose = (): void => {
        setOpen(false);
    };

    type RunTestState = {
        disableStart: boolean,
        testId: number | null,
        binaryId: number,
        envParams: string,
    }

    const [state, setState] = useState<RunTestState>({
        disableStart: true,
        testId: null,
        binaryId: 0,
        envParams: app === undefined ? '' : app?.DefaultParameter.replace(';', "\n"),
    });

    const [tests, setTests] = useState<ITestData[]>([]);

    useEffect(() => {
        if (appId !== null) {
            getAllTests(params.project_id as string, appId).then(response => {
                setTests(response.data);
            }).catch(e => {
                console.log(e);
            });
        }
    }, [params.project_id, appId]);

    const typeString = (type: number): string => {
        switch (type) {
            case 0:
                return 'Unity';
            case 1:
                return 'Cocos';
            case 2:
                return 'Serenity';
            case 3:
                return 'Scenario';
        }
        return '';
    };

    const executionString = (type: number): string => {
        switch (type) {
            case 0:
                return 'Concurrent';
            case 1:
                return 'Simultaneously';
        }
        return '';
    };

    const onRunTest = (): void => {
        if (appId !== null) {
            executeTest(params.project_id as string, appId, state.testId, state.binaryId, state.envParams).then(response => {
                navigate(`/project/${params.project_id}/app/${appId}/test/${ state.testId }/run/${ response.data.ID }`);
            }).catch(error => {
                console.log(error);
            });
        }
    };

    const getDevices = (test: ITestData): string => {
        if (test.TestConfig.AllDevices) {
            return 'all';
        }
        if (test.TestConfig.Devices !== null) {
            return test.TestConfig.Devices.length.toString();
        }
        return 'n/a';
    };

    const getTests = (test: ITestData): string => {
        if (test.TestConfig.Unity !== undefined && test.TestConfig.Unity !== null) {
            if (test.TestConfig.Unity?.RunAllTests) {
                return 'all';
            }
            if (test.TestConfig.Unity.UnityTestFunctions !== null) {
                return test.TestConfig.Unity.UnityTestFunctions.length.toString();
            }
        }

        return 'n/a';
    };

    const requiresApp = app?.Platform !== PlatformType.Editor

    const onBinarySelectionChanged = (app: IAppBinaryData): void => {
        setState(prevState => ({...prevState, binaryId: app.ID, disableStart: requiresApp && app.ID === 0}));
    };

    const onEnvParamsChanged = (event: ChangeEvent<HTMLInputElement>): void => {
        setState(prevState => ({...prevState, envParams: event.target.value}));
    };


    useEffect(() => {
        if (app !== null && app !== undefined) {
            setState(prevState => ({...prevState, disableStart: requiresApp && app.ID === 0, envParams: app.DefaultParameter.replace(";", "\n")}));
        }
    }, [app]);

    return (
        <div>
            <Dialog open={ open } onClose={ handleRunClose } aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Execute Test</DialogTitle>
                <DialogContent>
                    { requiresApp && (
                        <>
                            <DialogContentText>
                                Select an existing App to execute the tests.<br/>
                                Or Upload a new one.<br/>
                                <br/>
                            </DialogContentText>
                            <BinarySelection binaryId={state.binaryId} upload={ true } onSelectionChanged={ onBinarySelectionChanged }/>
                        </>
                    )}
                    You can change parameters of your app by providing key value pairs in an environment like format:<br/>
                    <br/>
                    <Typography variant={ 'subtitle2' }>
                        server=http://localhost:8080<br/>
                        user=autohub
                    </Typography>
                    <br/>
                    <TextField
                        label="Parameter"
                        fullWidth={ true }
                        multiline={ true }
                        rows={ 4 }
                        defaultValue={state.envParams}
                        value={state.envParams}
                        variant="outlined"
                        onChange={ onEnvParamsChanged }
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={ handleRunClose } color="primary">
                        Cancel
                    </Button>
                    <Button onClick={ () => {
                        onRunTest();
                        handleRunClose();
                    } } color="primary" variant={ 'contained' } disabled={ state.disableStart }>
                        Start
                    </Button>
                </DialogActions>
            </Dialog>
            <TableContainer component={ Paper }>
                <Table size="small" aria-label="a dense table">
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell align="right">Typ</TableCell>
                            <TableCell align="right">Execution</TableCell>
                            <TableCell align="right">Devices</TableCell>
                            <TableCell align="right">Tests</TableCell>
                            <TableCell align="right"/>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        { tests.map((test) => <TableRow key={ test.Name }>
                            <TableCell component="th" scope="row">{ test.Name }</TableCell>
                            <TableCell align="right">{ typeString(test.TestConfig.Type) }</TableCell>
                            <TableCell align="right">{ executionString(test.TestConfig.ExecutionType) }</TableCell>
                            <TableCell align="right">{ getDevices(test) }</TableCell>
                            <TableCell align="right">{ getTests(test) }</TableCell>
                            <TableCell align="right">
                                <ButtonGroup color="primary" aria-label="text button group">
                                    <Button variant="text" size="small"
                                        href={ `test/${ test.ID }` }>Show</Button>
                                    <Button variant="text" size="small"
                                        href={ `test/${ test.ID }/runs/last` }>Protocol</Button>
                                    <Button variant="text" size="small" endIcon={ <PlayArrow/> }
                                        onClick={ () => {
                                            setState(prevState => ({...prevState, testId: test.ID}))
                                            handleRunClickOpen();
                                        } }>Run</Button>
                                </ButtonGroup>


                            </TableCell>
                        </TableRow>) }
                    </TableBody>
                </Table>
            </TableContainer>
        </div>
    );
};

export default TestsTable;
