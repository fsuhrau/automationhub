import React from 'react';
import Paper from '@mui/material/Paper';
import {
    Alert,
    Box,
    Button,
    FormControl,
    FormControlLabel,
    Grid,
    InputLabel,
    Radio,
    RadioGroup,
    Select, SelectChangeEvent,
    Step,
    StepLabel,
    Stepper,
    TextField,
    Typography,
} from '@mui/material';
import { TestExecutionType } from '../../types/test.execution.type.enum';
import { TestType } from '../../types/test.type.enum';
import DeviceSelection from '../../components/device-selection.component';
import IDeviceData from '../../types/device';
import { useHistory } from 'react-router-dom';
import ICreateTestData from '../../types/request.create.test';
import { createTest } from '../../services/test.service';
import TestMethodSelection from '../../components/testmethod-selection.component';
import IAppFunctionData from '../../types/app.function';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';

const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: en[ key ] }));
}

function getExecutionTypes(): Array<Object> {
    return ToArray(TestExecutionType);
}

function getTestTypes(): Array<Object> {
    return ToArray(TestType);
}

function getSteps(): Array<string> {
    return ['Select Test Type', 'Test Configuration', 'Device Selection'];
}

function getUnityTestsConfig(): Array<Object> {
    return [{ id: 0, name: 'Run all Tests' }, { id: 1, name: 'Run only Selected Tests' }];
}

function getDeviceOption(): Array<Object> {
    return [{ id: 0, name: 'All Devices' }, { id: 1, name: 'Selected Devices Only' }];
}

const AddTestPage: React.FC = () => {

    const history = useHistory();

    const [activeStep, setActiveStep] = React.useState(0);
    const steps = getSteps();

    const testTypes = getTestTypes();
    const [testType, setTestType] = React.useState<TestType>(TestType.Unity);
    const handleTestTypeChange = (event: SelectChangeEvent<TestType>): void => {
        const type = (+event.target.value as TestType);
        setTestType(type);
    };

    const executionTypes = getExecutionTypes();
    const [executionType, setExecutionType] = React.useState<TestExecutionType>(TestExecutionType.Concurrent);
    const handleExecutionTypeChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        const type = (+event.target.value as TestExecutionType);
        setExecutionType(type);
    };

    const [testName, setTestName] = React.useState('');
    const handleTestNameChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        setTestName(event.target.value);
    };

    const unityTestExecutionTypes = getUnityTestsConfig();
    const [unityTestExecution, setUnityTestExecution] = React.useState<number>(0);
    const handleUnityTestExecutionChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        setUnityTestExecution(+event.target.value);
    };

    const deviceTypes = getDeviceOption();
    const [deviceType, setDeviceType] = React.useState<number>(0);
    const handleDeviceTypeChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        setDeviceType(+event.target.value);
    };

    const [selectedDevices, setSelectedDevices] = React.useState<IDeviceData[]>([]);
    const handleDeviceSelectionChanged = (devices: IDeviceData[]): void => {
        setSelectedDevices(devices);
    };

    const [unityTestFunctions, setUnityTestFunctions] = React.useState<IAppFunctionData[]>([]);

    const createNewTest = (): void => {
        const deviceIds: number[] = selectedDevices.map(value => value.ID) as number[];

        const requestData: ICreateTestData = {
            Name: testName,
            TestType: testType,
            ExecutionType: executionType,
            UnityAllTests: unityTestExecution === 0,
            UnitySelectedTests: unityTestFunctions,
            AllDevices: deviceType === 0,
            SelectedDevices: deviceIds,
        };

        createTest(requestData).then(response => {
            console.log(response.data);
            history.push('/web/tests');
        }).catch(ex => {
            console.log(ex);
        });
    };

    const handleNext = (): void => {
        if (activeStep === 2) {
            createNewTest();
        }
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
    };

    const handleBack = (): void => {
        setActiveStep((prevActiveStep) => prevActiveStep - 1);
    };

    const handleFunctionSelection = (funcs: IAppFunctionData[]): void => {
        setUnityTestFunctions(funcs);
    };

    return (
        <Paper sx={{ maxWidth: 1200, margin: 'auto', overflow: 'hidden' }}>
            <AppBar
                position="static"
                color="default"
                elevation={0}
                sx={{ borderBottom: '1px solid rgba(0, 0, 0, 0.12)' }}
            >
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            <Typography variant={ 'h6' }>
                                Create a new Test
                            </Typography>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={ { width: '100%', padding: 5 } }>
                <Stepper activeStep={ activeStep } alternativeLabel={ true }>
                    { steps.map((label) => (
                        <Step key={ label }>
                            <StepLabel>{ label }</StepLabel>
                        </Step>
                    )) }
                </Stepper>
                <Box sx={ { width: '100%', padding: 5 } }>
                    { activeStep === steps.length ? (
                        <div>
                            <Typography variant={'body1'}>Test is being created wait a moment and
                                you
                                get redirected</Typography>
                        </div>
                    ) : (
                        <div>
                            <Grid container={ true } spacing={ 5 }>
                                <Grid item={ true } xs={ 12 }>
                                    <Grid container={ true } justifyContent="center" spacing={ 5 }>
                                        <Grid item={ true }>
                                            { activeStep === 0 && (
                                                <Grid container={ true } justifyContent="center" spacing={ 2 }
                                                    alignItems={ 'center' } direction={ 'column' }>
                                                    <Grid item={ true } xs={ 6 }>
                                                        <FormControl>
                                                            <TextField required={ true } id="test-name" label="Name"
                                                                value={ testName }
                                                                onChange={ handleTestNameChange }/>
                                                        </FormControl>
                                                    </Grid>
                                                    <Grid item={ true } xs={ 6 }>
                                                        <FormControl>
                                                            <InputLabel htmlFor="test-type-selection">Test
                                                                Type</InputLabel>
                                                            <Select native={ true } value={ testType }
                                                                name={ 'test-type-selection' }
                                                                onChange={ handleTestTypeChange }
                                                                inputProps={ {
                                                                    name: 'Test Type',
                                                                    id: 'test-types',
                                                                } }>
                                                                <option aria-label="None" value=""
                                                                    key={ 'tt_none' }/>
                                                                { testTypes.map((value) => (
                                                                    <option key={ 'tt_' + value.id.toString() }
                                                                        value={ value.id.toString() }>{ value.name }</option>
                                                                )) }
                                                            </Select>
                                                        </FormControl>
                                                    </Grid>
                                                    <Grid item={ true } xs={ 6 }>
                                                        <Grid container={ true } spacing={ 2 }
                                                            alignItems={ 'center' }
                                                            direction={ 'row' }>
                                                            <Grid item={ true }>
                                                                <RadioGroup
                                                                    name="execution-type-selection"
                                                                    aria-label="spacing"
                                                                    value={ executionType.toString() }
                                                                    onChange={ handleExecutionTypeChange }
                                                                    row={ true }
                                                                >
                                                                    { executionTypes.map((value) => (
                                                                        <FormControlLabel
                                                                            key={ 'exec_' + value.id }
                                                                            value={ value.id.toString() }
                                                                            control={ <Radio/> }
                                                                            label={ value.name }
                                                                        />
                                                                    )) }
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
                                            ) }
                                            { activeStep === 1 && (
                                                <Grid container={ true } justifyContent="center" spacing={ 2 }
                                                    alignItems={ 'center' } direction={ 'column' }>
                                                    { testType === TestType.Unity && (
                                                        <>
                                                            <Grid item={ true } >
                                                                <RadioGroup
                                                                    name="unity-test-execution-selection"
                                                                    aria-label="spacing"
                                                                    value={ unityTestExecution.toString() }
                                                                    onChange={ handleUnityTestExecutionChange }
                                                                    row={ true }
                                                                >
                                                                    { unityTestExecutionTypes.map((value) => (
                                                                        <FormControlLabel
                                                                            key={ 'unityt_' + value.id }
                                                                            value={ value.id.toString() }
                                                                            control={ <Radio/> }
                                                                            label={ value.name }
                                                                        />
                                                                    )) }
                                                                </RadioGroup>
                                                            </Grid>
                                                            <Grid item={ true }>
                                                                { unityTestExecution === 1 && (
                                                                    <div>
                                                                        <TestMethodSelection onSelectionChanged={ handleFunctionSelection }/>
                                                                    </div>
                                                                ) }
                                                            </Grid>
                                                        </>
                                                    ) }
                                                </Grid>
                                            ) }
                                            { activeStep === 2 && (
                                                <Grid container={ true } justifyContent="center" spacing={ 2 }
                                                    alignItems={ 'center' } direction={ 'column' }>
                                                    <Grid item={ true }>
                                                        <RadioGroup
                                                            name="device-selection"
                                                            aria-label="spacing"
                                                            value={ deviceType.toString() }
                                                            onChange={ handleDeviceTypeChange }
                                                            row={ true }
                                                        >
                                                            { deviceTypes.map((value) => (
                                                                <FormControlLabel
                                                                    key={ 'device_' + value.id }
                                                                    value={ value.id.toString() }
                                                                    control={ <Radio/> }
                                                                    label={ value.name }
                                                                />
                                                            )) }
                                                        </RadioGroup>
                                                    </Grid>
                                                </Grid>
                                            ) }
                                            { activeStep === 2 && deviceType === 1 && (
                                                <Grid container={ true } justifyContent="center" spacing={ 2 }
                                                    alignItems={ 'center' } direction={ 'column' }>
                                                    <Grid item={ true }>
                                                        <Typography variant={ 'h6' }>
                                                            Select Devices
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={ true }>
                                                        <DeviceSelection selectedDevices={ selectedDevices }
                                                            onSelectionChanged={ handleDeviceSelectionChanged }/>
                                                    </Grid>
                                                </Grid>
                                            ) }
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Grid>
                        </div>
                    ) }
                </Box>
            </Box>
            <Grid container={ true } justifyContent={'flex-end'} >
                <Grid item={ true }>
                    <Box sx={ { p: 2, m: 2 } }>
                        <Button
                            disabled={ activeStep === 0 }
                            onClick={ handleBack }
                        >
                            Back
                        </Button>
                        <Button variant="contained" color="primary" onClick={ handleNext }>
                            { activeStep === steps.length - 1 ? 'Create' : 'Next' }
                        </Button>
                    </Box>
                </Grid>
            </Grid>
        </Paper>
    );
};

export default AddTestPage;
