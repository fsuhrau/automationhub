import React, {useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import {
    Alert,
    Box,
    Button,
    FormControl,
    FormControlLabel,
    Grid,
    InputLabel,
    MenuItem,
    Radio,
    RadioGroup,
    Select,
    Step,
    StepLabel,
    Stepper,
    TextField,
    Typography,
} from '@mui/material';
import {getExecutionTypes, TestExecutionType} from '../../types/test.execution.type.enum';
import {getTestTypes, TestType} from '../../types/test.type.enum';
import DeviceSelection from '../../components/device-selection.component';
import {useNavigate} from 'react-router-dom';
import ICreateTestData from '../../types/request.create.test';
import {createTest} from '../../services/test.service';
import TestMethodSelection from '../../components/testmethod-selection.component';
import IAppFunctionData from '../../types/app.function';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import {Delete} from '@mui/icons-material';
import Table from "@mui/material/Table";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import TableBody from "@mui/material/TableBody";
import IDeviceData from "../../types/device";
import {getAllDevices} from "../../services/device.service";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {ApplicationProps} from "../../application/ApplicationProps";

import {isEqual} from "lodash"
import {IdName} from "../../helper/enum_to_array";
import {getUnityTestCategoryTypes, UnityTestCategory} from "../../types/unity.test.category.type.enum";
import {useApplicationContext} from "../../hooks/ApplicationProvider";

function getSteps(): Array<string> {
    return ['Select Test Type', 'Test Configuration', 'Device Selection'];
}

export function getUnityTestsConfig(): Array<IdName> {
    return [{id: '0', name: 'Run all Tests'}, {id: '1', name: 'Run all of Category'}, {
        id: '2',
        name: 'Run only Selected Tests',
    }];
}

export function getDeviceOption(): Array<IdName> {
    return [{id: '0', name: 'All Devices'}, {id: '1', name: 'Selected Devices Only'}];
}

const AddTestPage: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {appState} = props;

    const {project, projectId} = useProjectContext();
    const {appId} = useApplicationContext();

    enum TestCreationSteps {
        Basics,
        Tests,
        Devices,
    }

    const steps = getSteps();
    const [activeStep, setActiveStep] = React.useState<TestCreationSteps>(TestCreationSteps.Basics);

    type NewTestState = {
        testName: string,
        testType: TestType,
        executionType: TestExecutionType,
        unityTestCategoryType: UnityTestCategory,
        deviceType: number,
        selectedDevices: number[],
        selectedTestFunctions: IAppFunctionData[],
        testCategories: string[],
        category: string,
    };

    const navigate = useNavigate();

    const [state, setState] = React.useState<NewTestState>({
            testType: TestType.Unity,
            executionType: TestExecutionType.Concurrent,
            testName: "",
            unityTestCategoryType: 0,
            deviceType: 0,
            selectedDevices: [],
            selectedTestFunctions: [],
            testCategories: [],
            category: '',
        }
    )


    const testTypes = getTestTypes();
    const executionTypes = getExecutionTypes();
    const unityTestCategoryTypes = getUnityTestCategoryTypes();
    const deviceTypes = getDeviceOption();

    const addCategory = () => {
        if (state.category != '') {
            setState(prevState => ({
                ...prevState,
                testCategories: [...prevState.testCategories, prevState.category],
                category: ''
            }))
        }
    };
    const removeCategory = (index: number) => {
        setState(prevState => ({
            ...prevState,
            testCategories: [...prevState.testCategories.slice(0, index), ...prevState.testCategories.slice(index + 1)]
        }))
    };

    const createNewTest = (): void => {

        const requestData: ICreateTestData = {
            Name: state.testName,
            TestType: state.testType,
            ExecutionType: state.executionType,
            UnityTestCategoryType: state.unityTestCategoryType,
            UnitySelectedTests: state.selectedTestFunctions,
            Categories: state.testCategories,
            AllDevices: state.deviceType === 0,
            SelectedDevices: state.selectedDevices,
        };

        createTest(projectId, appId, requestData).then(response => {
            navigate(`/project/${projectId}/app/tests`);
        }).catch(ex => {
            console.log(ex);
        });
    };

    const handleNext = (): void => {
        if (activeStep === TestCreationSteps.Devices) {
            createNewTest();
        }
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
    };

    const handleBack = (): void => {
        setActiveStep((prevActiveStep) => prevActiveStep - 1);
    };

    const [devices, setDevices] = useState<IDeviceData[]>([]);
    useEffect(() => {
        const app = project.Apps.find(a => a.ID === appId);
        getAllDevices(projectId, app?.Platform).then(response => {
            setDevices(response.data);
        })
    }, [project.Apps, projectId, appId])

    const onDeviceSelectionChanged = (selectedDevices: number[]) => {
        if (!isEqual(selectedDevices, state.selectedDevices)) {
            setState(prevState => ({...prevState, selectedDevices: selectedDevices}))
        }
    }

    const onTestSelectionChanged = (testSelection: IAppFunctionData[]) => {
        if (!isEqual(testSelection, state.selectedTestFunctions)) {
            setState(prevState => ({...prevState, selectedTestFunctions: testSelection}))
        }
    }

    return (
        <Paper sx={{maxWidth: 1200, margin: 'auto', overflow: 'hidden'}}>
            <AppBar
                position="static"
                color="default"
                elevation={0}
                sx={{borderBottom: '1px solid rgba(0, 0, 0, 0.12)'}}
            >
                <Toolbar>
                    <Grid container={true} spacing={2} alignItems="center">
                        <Grid item={true}>
                            <Typography variant={'h6'}>
                                Create a new Test
                            </Typography>
                        </Grid>
                        <Grid item={true} xs={true}>
                        </Grid>
                        <Grid item={true}>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={{width: '100%', padding: 5}}>
                <Stepper activeStep={activeStep} alternativeLabel={true}>
                    {steps.map((label) => (
                        <Step key={label}>
                            <StepLabel>{label}</StepLabel>
                        </Step>
                    ))}
                </Stepper>
                <Box sx={{width: '100%', padding: 5}}>
                    {activeStep === steps.length ? (
                        <div>
                            <Typography variant={'body1'}>Test is being created wait a moment and
                                you
                                get redirected</Typography>
                        </div>
                    ) : (
                        <div>
                            <Grid container={true} spacing={5}>
                                <Grid item={true} xs={12}>
                                    <Grid container={true} justifyContent="center" spacing={5}>
                                        <Grid item={true}>
                                            {activeStep === TestCreationSteps.Basics && (
                                                <Grid container={true} justifyContent="center" spacing={1}
                                                      alignItems={'center'}>
                                                    <Grid item={true} xs={6}>
                                                        <FormControl fullWidth={true}>
                                                            <TextField required={true} id="test-name" label="Name"
                                                                       value={state.testName}
                                                                       onChange={event => setState(prevState => ({
                                                                           ...prevState,
                                                                           testName: event.target.value
                                                                       }))}/>
                                                        </FormControl>
                                                    </Grid>
                                                    <Grid item={true} xs={12}/>
                                                    <Grid item={true} xs={2}>
                                                        <FormControl fullWidth={true}>
                                                            <InputLabel id="test-type-selection">Test Type</InputLabel>
                                                            <Select
                                                                defaultValue={state.testType}
                                                                labelId="test-type-selection"
                                                                id="test-type"
                                                                label="Test Type"
                                                                onChange={event => setState(prevState => ({
                                                                    ...prevState,
                                                                    testType: +event.target.value as TestType
                                                                }))}
                                                            >
                                                                {testTypes.map((value) => (
                                                                    <MenuItem key={'tt_' + value.id}
                                                                              value={value.id}>{value.name}</MenuItem>
                                                                ))}
                                                            </Select>
                                                        </FormControl>
                                                    </Grid>
                                                    <Grid item={true} xs={12}/>
                                                    <Grid item={true} xs={6}>
                                                        <Grid container={true} spacing={2}
                                                              alignItems={'center'}
                                                              direction={'row'}>
                                                            <Grid item={true}>
                                                                <RadioGroup
                                                                    name="execution-type-selection"
                                                                    aria-label="spacing"
                                                                    value={state.executionType.toString()}
                                                                    onChange={event => setState(prevState => ({
                                                                        ...prevState,
                                                                        executionType: +event.target.value as TestExecutionType
                                                                    }))}
                                                                    row={true}
                                                                >
                                                                    {executionTypes.map((value) => (
                                                                        <FormControlLabel
                                                                            key={'exec_' + value.id}
                                                                            value={value.id}
                                                                            control={<Radio/>}
                                                                            label={value.name}
                                                                        />
                                                                    ))}
                                                                </RadioGroup>
                                                                <Alert severity="info">
                                                                    Concurrent = runs each test on a different free
                                                                    device to get faster results<br/>
                                                                    Simultaneously = runs every test on every device
                                                                    to get a better accuracy</Alert>
                                                            </Grid>
                                                        </Grid>
                                                    </Grid>
                                                </Grid>
                                            )}
                                            {activeStep === TestCreationSteps.Tests && (
                                                <Grid container={true} justifyContent="center" spacing={2}
                                                      alignItems={'center'} direction={'column'}>
                                                    {state.testType === TestType.Unity && (
                                                        <>
                                                            <Grid item={true}>
                                                                <RadioGroup
                                                                    name="unity-test-execution-selection"
                                                                    aria-label="spacing"
                                                                    value={state.unityTestCategoryType.toString()}
                                                                    onChange={event => setState(prevState => ({
                                                                        ...prevState,
                                                                        unityTestCategoryType: +event.target.value
                                                                    }))}
                                                                    row={true}
                                                                >
                                                                    {unityTestCategoryTypes.map((value) => (
                                                                        <FormControlLabel
                                                                            key={'unityt_' + value.id}
                                                                            value={value.id.toString()}
                                                                            control={<Radio/>}
                                                                            label={value.name}
                                                                        />
                                                                    ))}
                                                                </RadioGroup>
                                                            </Grid>
                                                            <Grid item={true}>
                                                                {state.unityTestCategoryType === UnityTestCategory.RunAllOfCategory && (
                                                                    <Grid container={true} justifyContent="center"
                                                                          spacing={1} alignItems={'center'}>
                                                                        <Grid item={true}>
                                                                            <Table size="small"
                                                                                   aria-label="a dense table"
                                                                                   width={600}>
                                                                                <TableHead>
                                                                                    <TableRow>
                                                                                        <TableCell>Category</TableCell>
                                                                                        <TableCell></TableCell>
                                                                                    </TableRow>
                                                                                </TableHead>
                                                                                <TableBody>
                                                                                    {state.testCategories.map((element, index) =>
                                                                                        <TableRow
                                                                                            key={`categories_${index}`}>
                                                                                            <TableCell>{element}</TableCell>
                                                                                            <TableCell>
                                                                                                <IconButton edge="end"
                                                                                                            aria-label="delete"
                                                                                                            onClick={() => removeCategory(index)}>
                                                                                                    <Delete/>
                                                                                                </IconButton>
                                                                                            </TableCell>
                                                                                        </TableRow>)}

                                                                                    <TableRow>
                                                                                        <TableCell>
                                                                                            <FormControl>
                                                                                                <TextField
                                                                                                    required={true}
                                                                                                    id="test-name"
                                                                                                    label="Category Name"
                                                                                                    value={state.category}
                                                                                                    fullWidth={true}
                                                                                                    onChange={event => setState(prevState => ({
                                                                                                        ...prevState,
                                                                                                        category: event.target.value
                                                                                                    }))}
                                                                                                />
                                                                                            </FormControl>
                                                                                        </TableCell>
                                                                                        <TableCell>
                                                                                            <Button variant={'outlined'}
                                                                                                    onClick={addCategory}>Add</Button>
                                                                                        </TableCell>
                                                                                    </TableRow>
                                                                                </TableBody>
                                                                            </Table>
                                                                        </Grid>
                                                                    </Grid>
                                                                )}
                                                                {state.unityTestCategoryType === UnityTestCategory.RunSelectedTestsOnly && (
                                                                    <TestMethodSelection
                                                                        onSelectionChanged={onTestSelectionChanged}/>
                                                                )}
                                                            </Grid>
                                                        </>
                                                    )}
                                                </Grid>
                                            )}
                                            {activeStep === TestCreationSteps.Devices && (
                                                <Grid container={true} justifyContent="center" spacing={2}
                                                      alignItems={'center'} direction={'column'}>
                                                    <Grid item={true}>
                                                        <RadioGroup
                                                            name="device-selection"
                                                            aria-label="spacing"
                                                            value={state.deviceType.toString()}
                                                            onChange={event => setState(prevState => ({
                                                                ...prevState,
                                                                deviceType: +event.target.value
                                                            }))}
                                                            row={true}
                                                        >
                                                            {deviceTypes.map((value) => (
                                                                <FormControlLabel
                                                                    key={'device_' + value.id}
                                                                    value={value.id}
                                                                    control={<Radio/>}
                                                                    label={value.name}
                                                                />
                                                            ))}
                                                        </RadioGroup>
                                                    </Grid>
                                                </Grid>
                                            )}
                                            {activeStep === TestCreationSteps.Devices && state.deviceType === 1 && (
                                                <Grid container={true} justifyContent="center" spacing={2}
                                                      alignItems={'center'} direction={'column'}>
                                                    <Grid item={true}>
                                                        <Typography variant={'h6'}>
                                                            Select Devices
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={true}>
                                                        {
                                                            devices !== undefined && devices.length > 0 &&
                                                            <DeviceSelection
                                                                devices={devices}
                                                                selectedDevices={state.selectedDevices}
                                                                onSelectionChanged={onDeviceSelectionChanged}/>
                                                        }
                                                    </Grid>
                                                </Grid>
                                            )}
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Grid>
                        </div>
                    )}
                </Box>
            </Box>
            <Grid container={true} justifyContent={'flex-end'}>
                <Grid item={true}>
                    <Box sx={{p: 2, m: 2}}>
                        <Button
                            disabled={activeStep === 0}
                            onClick={handleBack}
                        >
                            Back
                        </Button>
                        <Button variant="contained" color="primary" onClick={handleNext}>
                            {activeStep === steps.length - 1 ? 'Create' : 'Next'}
                        </Button>
                    </Box>
                </Grid>
            </Grid>
        </Paper>
    );
};

export default AddTestPage;
