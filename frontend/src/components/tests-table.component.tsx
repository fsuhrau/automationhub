import React, {ChangeEvent, useEffect, useState} from 'react';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import ITestData from '../types/test';
import {executeTest, getAllTests} from '../services/test.service';
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
import {PlayArrow} from '@mui/icons-material';
import BinarySelection from './binary-selection.component';
import {useNavigate} from 'react-router-dom';
import {PlatformType} from '../types/platform.type.enum';
import {IAppBinaryData} from "../types/app";
import {ApplicationProps} from "../application/ApplicationProps";
import {useProjectContext} from "../hooks/ProjectProvider";
import {DataGrid, GridColDef, GridRowsProp} from "@mui/x-data-grid";
import Chip from "@mui/material/Chip";

function renderChip(type: string) {
    return <Chip label={type} color={'default'} size="small"/>;
}

function renderActions(id: number) {
    return <ButtonGroup color="primary" aria-label="text button group">
        <Button variant="text" size="small"
                href={`test/${id}`}>Show</Button>
        <Button variant="text" size="small"
                href={`test/${id}/runs/last`}>Protocol</Button>
        <Button variant="text" size="small" endIcon={<PlayArrow/>}
                onClick={() => {
                    setState(prevState => ({...prevState, testId: id}))
                    handleRunClickOpen();
                }}>Run</Button>
    </ButtonGroup>;
}

export const columns: GridColDef[] = [
    {
        field: 'name',
        headerName: 'Name',
        flex: 1.5,
        minWidth: 300
    },
    {
        field: 'type',
        headerName: 'Type',
        flex: 0.5,
        minWidth: 90,
        renderCell: (params) => renderChip(params.value as any),
    },
    {
        field: 'execution',
        headerName: 'Execution',
        flex: 0.5,
        minWidth: 90,
        renderCell: (params) => renderChip(params.value as any),
    },
    {
        field: 'devices',
        headerName: 'Devices',
        headerAlign: 'right',
        align: 'right',
        flex: 1,
        minWidth: 100,
    },
    {
        field: 'tests',
        headerName: 'Tests',
        headerAlign: 'right',
        align: 'right',
        flex: 0.5,
        minWidth: 90,
    },
    {
        field: 'actions',
        headerName: '',
        headerAlign: 'right',
        align: 'right',
        flex: 1,
        minWidth: 100,
        renderCell: (params) => renderActions(params.value as any),
    },
];

interface TestTableProps extends ApplicationProps {
    appId: number | null
}

const TestsTable: React.FC<TestTableProps> = (props: TestTableProps) => {

    const {appId, appState} = props;
    const {project, projectId} = useProjectContext();

    const navigate = useNavigate();

    const app = project.Apps.find(a => a.ID === appId);

    // dialog
    const [open, setOpen] = useState(false);
    const [gridData, setGridData] = useState<GridRowsProp[]>([]);

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

    useEffect(() => {
        if (appId !== null) {
            getAllTests(projectId, appId).then(response => {
                setGridData(response.data.map(d => {
                    return {
                        id: d.ID,
                        name: d.Name,
                        type: typeString(d.TestConfig.Type),
                        execution: executionString(d.TestConfig.ExecutionType),
                        devices: getDevices(d),
                        tests: getTests(d),
                        actions: d.ID,
                    }
                }));
            }).catch(e => {
                console.log(e);
            });
        }
    }, [projectId, appId]);

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
            executeTest(projectId as string, appId, state.testId, state.binaryId, state.envParams).then(response => {
                navigate(`/project/${projectId}/app/test/${state.testId}/run/${response.data.ID}`);
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
            setState(prevState => ({
                ...prevState,
                disableStart: requiresApp && app.ID === 0,
                envParams: app.DefaultParameter.replace(";", "\n")
            }));
        }
    }, [app]);

    return (
        <>
            <Dialog open={open} onClose={handleRunClose} aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Execute Test</DialogTitle>
                <DialogContent>
                    {requiresApp && (
                        <>
                            <DialogContentText>
                                Select an existing App to execute the tests.<br/>
                                Or Upload a new one.<br/>
                                <br/>
                            </DialogContentText>
                            <BinarySelection binaryId={state.binaryId} upload={true}
                                             onSelectionChanged={onBinarySelectionChanged}/>
                        </>
                    )}
                    You can change parameters of your app by providing key value pairs in an environment like
                    format:<br/>
                    <br/>
                    <Typography variant={'subtitle2'}>
                        server=http://localhost:8080<br/>
                        user=autohub
                    </Typography>
                    <br/>
                    <TextField
                        label="Parameter"
                        fullWidth={true}
                        multiline={true}
                        rows={4}
                        defaultValue={state.envParams}
                        value={state.envParams}
                        variant="outlined"
                        onChange={onEnvParamsChanged}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleRunClose} color="primary">
                        Cancel
                    </Button>
                    <Button onClick={() => {
                        onRunTest();
                        handleRunClose();
                    }} color="primary" variant={'contained'} disabled={state.disableStart}>
                        Start
                    </Button>
                </DialogActions>
            </Dialog>
            <DataGrid
                autoHeight
                disableRowSelectionOnClick
                rows={gridData}
                columns={columns}
                getRowClassName={(params) =>
                    params.indexRelativeToCurrentPage % 2 === 0 ? 'even' : 'odd'
                }
                initialState={{
                    pagination: {paginationModel: {pageSize: 20}},
                }}
                pageSizeOptions={[10, 20, 50]}
                disableColumnResize
                density="compact"
                slotProps={{
                    filterPanel: {
                        filterFormProps: {
                            logicOperatorInputProps: {
                                variant: 'outlined',
                                size: 'small',
                            },
                            columnInputProps: {
                                variant: 'outlined',
                                size: 'small',
                                sx: {mt: 'auto'},
                            },
                            operatorInputProps: {
                                variant: 'outlined',
                                size: 'small',
                                sx: {mt: 'auto'},
                            },
                            valueInputProps: {
                                InputComponentProps: {
                                    variant: 'outlined',
                                    size: 'small',
                                },
                            },
                        },
                    },
                }}
            />
        </>
    );
};

export default TestsTable;
