import React, { ChangeEvent, useEffect, useState } from 'react';

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

interface TestTableProps {
    appId: number
}

const TestsTable: React.FC<TestTableProps> = (props: TestTableProps) => {

    const { appId } = props;

    let params = useParams();

    const navigate = useNavigate();

    // dialog
    const [open, setOpen] = useState(false);
    const handleRunClickOpen = (): void => {
        setOpen(true);
    };
    const handleRunClose = (): void => {
        setOpen(false);
    };

    // test handling
    const [startDisabled, setStartDisabled] = useState<boolean>(true);
    const [selectedTest, setSelectedTest] = useState<ITestData | null>(null);
    const [selectedBinaryID, setSelectedBinaryID] = useState<number>(0);
    const [envParameter, setEnvParameter] = useState<string>('');
    const [tests, setTests] = useState<ITestData[]>([]);

    useEffect(() => {
        getAllTests(params.project_id as string, appId).then(response => {
            setTests(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

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

    const onRunTest = (id: number, binaryId: number): void => {
        executeTest(params.project_id as string, appId, id, binaryId, envParameter).then(response => {
            navigate(`/web/test/${ id }/run/${ response.data.ID }`);
        }).catch(error => {
            console.log(error);
        });
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

    const onBinarySelectionChanged = (app: IAppBinaryData): void => {
        setSelectedBinaryID(app.ID);
    };

    const onEnvParamsChanged = (event: ChangeEvent<HTMLInputElement>): void => {
        setEnvParameter(event.target.value);
    };

    useEffect(() => {
        // selectedTest?.TestConfig.Platform == PlatformType.Editor
        setStartDisabled(!(false || selectedBinaryID > 0))
    }, [selectedTest, selectedBinaryID]);

    return (
        <div>
            <Dialog open={ open } onClose={ handleRunClose } aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Execute Test</DialogTitle>
                <DialogContent>
                    { /* editor has no platform selectedTest.TestConfig.Platform != PlatformType.Editor */}
                    {selectedTest && true && (
                        <>
                            <DialogContentText>
                                Select an existing App to execute the tests.<br/>
                                Or Upload a new one.<br/>
                                <br/>
                            </DialogContentText>
                            <BinarySelection appid={appId} upload={ true } onSelectionChanged={ onBinarySelectionChanged }/>
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
                        id="outlined-multiline-static"
                        label="Parameter"
                        fullWidth={ true }
                        multiline={ true }
                        rows={ 4 }
                        defaultValue=""
                        variant="outlined"
                        onChange={ onEnvParamsChanged }
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={ handleRunClose } color="primary">
                        Cancel
                    </Button>
                    <Button onClick={ () => {
                        selectedTest && onRunTest(selectedTest.ID, selectedBinaryID);
                        handleRunClose();
                    } } color="primary" variant={ 'contained' } disabled={ startDisabled }>
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
                                            setSelectedTest(test);
                                            //setSelectedTestID(test.ID as number);
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
