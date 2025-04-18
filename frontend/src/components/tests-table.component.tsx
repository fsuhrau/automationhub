import React, {ChangeEvent, useEffect, useState} from 'react';

import ITestData from '../types/test';
import {executeTest, getAllTests} from '../services/test.service';
import {
    Button,
    ButtonGroup,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
    Typography,
} from '@mui/material';
import {PlayArrow} from '@mui/icons-material';
import BinarySelection from './binary-selection.component';
import {useNavigate} from 'react-router-dom';
import {PlatformType} from '../types/platform.type.enum';
import {IAppBinaryData} from "../types/app";
import {useProjectContext} from "../hooks/ProjectProvider";
import {DataGrid, GridColDef} from "@mui/x-data-grid";
import Chip from "@mui/material/Chip";
import {UnityTestCategory} from "../types/unity.test.category.type.enum";
import Grid from "@mui/material/Grid";
import {getTestTypeName} from "../types/test.type.enum";
import {getTestExecutionName} from "../types/test.execution.type.enum";
import {useError} from "../ErrorProvider";
import TextareaAutosize from '@mui/material/TextareaAutosize';

interface TestTableProps {
    appId: number | null
}

const TestsTable: React.FC<TestTableProps> = (props: TestTableProps) => {

    const {appId} = props;
    const {project, projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const navigate = useNavigate();

    const app = project.Apps.find(a => a.ID === appId);

    // dialog
    const [open, setOpen] = useState(false);
    const [gridData, setGridData] = useState<{
        id: number,
        name: string,
        type: string,
        execution: string,
        devices: string,
        tests: string,
        actions: number
    }[]>([]);

    const handleRunClickOpen = (): void => {
        setOpen(true);
    };
    const handleRunClose = (): void => {
        setOpen(false);
    };

    type RunTestState = {
        testId: number | null,
        binaryId: number | null,
        envParams: string,
        startURL: string | null,
    }

    const [state, setState] = useState<RunTestState>({
        testId: null,
        binaryId: null,
        startURL: null,
        envParams: app === undefined ? '' : app?.DefaultParameter.replaceAll(';', '\n'),
    });


    const renderChip = (type: string) => {
        return <Chip label={type} color={'default'} size="small"/>;
    }

    const navigateAction = (route: string) => {
        navigate(route)
    }

    const renderActions = (id: number) => {
        return <ButtonGroup variant={"text"} aria-label="text button group">
            <Button size="small"
                    onClick={() => navigateAction(`/project/${projectIdentifier}/app:${appId}/test/${id}`)}>Show</Button>
            <Button size="small"
                    onClick={() => navigateAction(`/project/${projectIdentifier}/app:${appId}/test/${id}/runs/last`)}>Protocol</Button>
            <Button size="small" endIcon={<PlayArrow/>} onClick={() => {
                setState(prevState => ({...prevState, testId: id}))
                handleRunClickOpen();
            }}>Run</Button>
        </ButtonGroup>;
    }

    const columns: GridColDef[] = [
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

    useEffect(() => {
        if (appId !== null) {
            getAllTests(projectIdentifier, appId).then(response => {
                setGridData(response.data.map(d => {
                    return {
                        id: d.ID,
                        name: d.Name,
                        type: getTestTypeName(d.TestConfig.Type),
                        execution: getTestExecutionName(d.TestConfig.ExecutionType),
                        devices: getDevices(d),
                        tests: getTests(d),
                        actions: d.ID,
                    }
                }));
            }).catch(e => {
                setError(e);
            });
        }
    }, [projectIdentifier, appId]);

    const onRunTest = (): void => {
        if (appId !== null) {
            executeTest(projectIdentifier, appId, state.testId, {
                AppBinaryID: state.binaryId!,
                Params: state.envParams,
                StartURL: state.startURL,
            }).then(response => {
                navigate(`/project/${projectIdentifier}/app:${appId}/test/${state.testId}/run/${response.data.ID}`);
            }).catch(error => {
                setError(error);
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
            if (test.TestConfig.Unity?.UnityTestCategoryType == UnityTestCategory.RunAllTests) {
                return 'all';
            }
            if (test.TestConfig.Unity.UnityTestFunctions !== null) {
                return test.TestConfig.Unity.UnityTestFunctions.length.toString();
            }
        }

        return 'n/a';
    };

    const requiresApp = app?.Platform !== PlatformType.Editor && app?.Platform !== PlatformType.Web;
    const requiresURL = app?.Platform === PlatformType.Web;

    const onBinarySelectionChanged = (binary: IAppBinaryData | null): void => {
        setState(prevState => ({...prevState, binaryId: binary ? binary.ID : null}));
    };

    const onEnvParamsChanged = (event: ChangeEvent<HTMLTextAreaElement>): void => {
        setState(prevState => ({...prevState, envParams: event.target.value}));
    };


    useEffect(() => {
        if (app !== null && app !== undefined) {
            setState(prevState => ({
                ...prevState,
                envParams: app.DefaultParameter.replaceAll(";", "\n")
            }));
        }
    }, [app]);

    return (
        <div style={{display: 'flex', flexDirection: 'column'}}>
            <Dialog fullWidth={true}
                    maxWidth={"sm"}
                    open={open} onClose={handleRunClose} aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Execute Test</DialogTitle>
                <DialogContent>
                    <Grid container={true} spacing={2}>
                        {requiresApp && (
                            <Grid container={true} size={12} spacing={2}>
                                <Grid size={12}>
                                    <Typography variant={"body1"}>
                                        Select an existing App to execute the tests.<br/>
                                        Or Upload a new one.<br/>
                                    </Typography>
                                </Grid>
                                <Grid size={12}>
                                    <BinarySelection binaryId={state.binaryId} upload={true}
                                                     onSelectionChanged={onBinarySelectionChanged}/>
                                </Grid>
                            </Grid>
                        )}
                        {requiresURL && <Grid size={12}>
                            <Grid size={12} container={true} spacing={2}>
                                <Typography variant={"body1"}>
                                    Set the Startup URL for your test<br/>
                                </Typography>
                            </Grid>
                            <Grid size={12}>
                                <TextField required={true} fullWidth={true} value={state.startURL}
                                           onChange={(e) => setState(prevState => ({
                                               ...prevState,
                                               startURL: e.target.value
                                           }))}/>
                            </Grid>
                        </Grid>}
                        <Grid size={12}>
                            <Typography variant={"body1"}>
                                You can change parameters of your app by providing key value pairs in an environment
                                like
                                format:
                            </Typography>
                        </Grid>
                        <Grid size={12}>
                            <Typography variant={'subtitle2'}>
                                server=http://localhost:8080<br/>
                                user=autohub
                            </Typography>
                        </Grid>
                        <Grid size={12}>
                            <TextareaAutosize
                                style={{width: '100%', height: '100px'}}
                                placeholder={"Parameter"}
                                defaultValue={state.envParams}
                                value={state.envParams}
                                onChange={onEnvParamsChanged}
                            />

                        </Grid>
                    </Grid>
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleRunClose} color="primary">
                        Cancel
                    </Button>
                    <Button onClick={() => {
                        onRunTest();
                        handleRunClose();
                    }} color="primary" variant={'contained'} disabled={requiresApp && state.binaryId === null}>
                        Start
                    </Button>
                </DialogActions>
            </Dialog>
            <DataGrid
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
        </div>
    );
};

export default TestsTable;
