import React, {useEffect, useState} from 'react';

import ITestData from '../types/test';
import {executeTest, getAllTests} from '../services/test.service';
import {
    Button,
    ButtonGroup,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    MenuItem,
    Select,
    TextField,
    Typography,
} from '@mui/material';
import {PlayArrow} from '@mui/icons-material';
import BinarySelection from './binary-selection.component';
import {useNavigate} from 'react-router-dom';
import {PlatformType} from '../types/platform.type.enum';
import {AppParameterOption, IAppBinaryData} from "../types/app";
import {useProjectContext} from "../hooks/ProjectProvider";
import {DataGrid, GridColDef} from "@mui/x-data-grid";
import Chip from "@mui/material/Chip";
import {UnityTestCategory} from "../types/unity.test.category.type.enum";
import Grid from "@mui/material/Grid";
import {getTestTypeName} from "../types/test.type.enum";
import {getTestExecutionName} from "../types/test.execution.type.enum";
import {useError} from "../ErrorProvider";

interface TestTableProps {
    appId: number | null
}

const TestsTable: React.FC<TestTableProps> = (props: TestTableProps) => {

    const {appId} = props;
    const {project, projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const navigate = useNavigate();

    const app = project.apps.find(a => a.id === appId);

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

    interface Parameter {
        name: string
        value: string
    }

    type RunTestState = {
        testId: number | null,
        binaryId: number | null,
        startURL: string | null,
        envParams: Parameter[],
    }

    const [testRunState, setTestRunState] = useState<RunTestState>({
        testId: null,
        binaryId: null,
        startURL: null,
        envParams: app!.parameter.map(p => ({name: p.name, value: p.type.defaultValue})),
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
                setTestRunState(prevState => ({...prevState, testId: id}))
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
            getAllTests(projectIdentifier, appId).then(tests => {
                setGridData(tests.map(d => {
                    return {
                        id: d.id,
                        name: d.name,
                        type: getTestTypeName(d.testConfig.type),
                        execution: getTestExecutionName(d.testConfig.executionType),
                        devices: getDevices(d),
                        tests: getTests(d),
                        actions: d.id,
                    }
                }));
            }).catch(e => {
                setError(e);
            });
        }
    }, [projectIdentifier, appId]);

    const onRunTest = (): void => {
        if (appId !== null) {
            executeTest(projectIdentifier, appId, testRunState.testId, {
                appBinaryId: testRunState.binaryId!,
                params: testRunState.envParams.filter(p => p.value.length > 0).map(p => {
                    return `${p.name}=${p.value}`
                }).join(";"),
                startUrl: testRunState.startURL,
            }).then(testRun => {
                navigate(`/project/${projectIdentifier}/app:${appId}/test/${testRunState.testId}/run/${testRun.id}`);
            }).catch(error => {
                setError(error);
            });
        }
    };

    const getDevices = (test: ITestData): string => {
        if (test.testConfig.allDevices) {
            return 'all';
        }
        if (test.testConfig.devices !== null) {
            return test.testConfig.devices.length.toString();
        }
        return 'n/a';
    };

    const getTests = (test: ITestData): string => {
        if (test.testConfig.unity !== undefined && test.testConfig.unity !== null) {
            if (test.testConfig.unity?.testCategoryType == UnityTestCategory.RunAllTests) {
                return 'all';
            }
            if (test.testConfig.unity.testFunctions !== null) {
                return test.testConfig.unity.testFunctions.length.toString();
            }
        }

        return 'n/a';
    };

    const requiresApp = app?.platform !== PlatformType.Editor && app?.platform !== PlatformType.Web;
    const requiresURL = app?.platform === PlatformType.Web;

    const onBinarySelectionChanged = (binary: IAppBinaryData | null): void => {
        setTestRunState(prevState => ({...prevState, binaryId: binary ? binary.id : null}));
    };

    const handleParameterChange = (idx: number, value: string) => {
        setTestRunState(prevState => ({
            ...prevState,
            envParams: prevState.envParams.map((d, i) => idx === i ? {...d, value: value} : d)
        }));
    }

    useEffect(() => {
        if (app !== null && app !== undefined) {
            setTestRunState(prevState => ({
                ...prevState,
                envParams: app.parameter?.map(value => ({name: value.name, value: value.type.defaultValue}))
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
                                    <BinarySelection binaryId={testRunState.binaryId} upload={true}
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
                                <TextField required={true} fullWidth={true} value={testRunState.startURL}
                                           onChange={(e) => setTestRunState(prevState => ({
                                               ...prevState,
                                               startURL: e.target.value
                                           }))}/>
                            </Grid>
                        </Grid>}
                        <Grid size={12} container={true}>
                            {app?.parameter?.map((p, i) => {
                                if (p.type.type === 'string') return (
                                    <>
                                        <Grid size={4}>
                                            {p.name}
                                        </Grid>
                                        <Grid size={8}>
                                            <TextField defaultValue={p.type.defaultValue}
                                                       fullWidth={true}
                                                       value={testRunState.envParams[i].value}
                                                       onChange={e => handleParameterChange(i, e.target.value)}/>
                                        </Grid>
                                    </>
                                );
                                if (p.type.type === 'option') return (
                                    <>
                                        <Grid size={4}>
                                            {p.name}
                                        </Grid>
                                        <Grid size={8}>
                                            <Select
                                                id={`${p.name}-option-select`}
                                                value={testRunState.envParams[i].value}
                                                label={p.name}
                                                onChange={e => handleParameterChange(i, e.target.value)}>
                                                {
                                                    (p.type as AppParameterOption).options.map(o => (
                                                        <MenuItem value={o}>{o}</MenuItem>))
                                                }
                                            </Select>
                                        </Grid>
                                    </>);
                                return null;
                            })}
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
                    }} color="primary" variant={'contained'} disabled={requiresApp && testRunState.binaryId === null}>
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
