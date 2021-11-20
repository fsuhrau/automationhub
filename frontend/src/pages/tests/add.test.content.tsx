import React, { FC } from 'react';
import Paper from '@material-ui/core/Paper';
import { createStyles, Theme, withStyles } from '@material-ui/core/styles';
import {
    Button,
    FormControl,
    FormControlLabel,
    Grid,
    InputLabel,
    makeStyles,
    Radio,
    RadioGroup,
    Select,
    Step,
    StepLabel,
    Stepper,
    TextField,
    Typography,
} from '@material-ui/core';
import { TestExecutionType } from '../../types/test.execution.type.enum';
import { TestType } from '../../types/test.type.enum';
import DeviceSelection from '../../components/device-selection.component';
import IDeviceData from '../../types/device';
import { useHistory } from 'react-router-dom';
import ICreateTestData from '../../types/request.create.test';
import { createTest } from '../../services/test.service';
import TestMethodSelection from '../../components/testmethod-selection.component';
import IAppFunctionData from '../../types/app.function';

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        paper: {
            maxWidth: 1200,
            padding: '10px',
            overflow: 'hidden',
        },
        root: {
            width: '100%',
        },
        backButton: {
            marginRight: theme.spacing(1),
        },
        instructions: {
            marginTop: theme.spacing(1),
            marginBottom: theme.spacing(1),
        },
        formControl: {
            margin: theme.spacing(1),
            minWidth: 120,
        },
        selectEmpty: {
            marginTop: theme.spacing(2),
        },
    }),
);
const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({id: key, name: en[ key ]}));
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
    return [{id: 0, name: 'Run all Tests'}, {id: 1, name: 'Run only Selected Tests'}];
}

function getDeviceOption(): Array<Object> {
    return [{id: 0, name: 'All Devices'}, {id: 1, name: 'Selected Devices Only'}];
}

const AddTestPage: FC = () => {
    const classes = useStyles();
    const history = useHistory();
    const [activeStep, setActiveStep] = React.useState(0);
    const steps = getSteps();

    const testTypes = getTestTypes();
    const [testType, setTestType] = React.useState<TestType>(TestType.Unity);
    const handleTestTypeChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        const type = (+event.target.value as TestType);
        setTestType(type);
    };

    const executionTypes = getExecutionTypes();
    const [executionType, setExecutionType] = React.useState<TestExecutionType>(TestExecutionType.Concurrent);
    const handleExecutionTypeChange = (event: React.ChangeEvent<{ name?: string; value: TestExecutionType }>): void => {
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
        <div>
            <Typography variant={ "h4" }>
                Create a new Test
            </Typography>
            <br/>
            <Paper className={ classes.paper }>
                <div className={ classes.root }>
                    <Stepper activeStep={ activeStep } alternativeLabel={ true }>
                        { steps.map((label) => (
                            <Step key={ label }>
                                <StepLabel>{ label }</StepLabel>
                            </Step>
                        )) }
                    </Stepper>
                    <div>
                        { activeStep === steps.length ? (
                            <div>
                                <Typography className={ classes.instructions }>Test is being created wait a moment and
                                    you
                                    get redirected</Typography>
                            </div>
                        ) : (
                            <div>
                                <Grid container={ true } className={ classes.root } spacing={ 2 }>
                                    <Grid item={ true } xs={ 12 }>
                                        <Grid container={ true } justifyContent="center" spacing={ 2 }>
                                            <Grid item={ true }>
                                                { activeStep === 0 && (
                                                    <Grid container={ true } justifyContent="center" spacing={ 2 }
                                                          alignItems={ 'center' } direction={ 'column' }>
                                                        <Grid item={ true } sx={ 6 }>
                                                            <FormControl className={ classes.formControl }>
                                                                <TextField required={ true } id="test-name" label="Name"
                                                                           value={ testName }
                                                                           onChange={ handleTestNameChange }/>
                                                            </FormControl>
                                                        </Grid>
                                                        <Grid item={ true } sx={ 6 }>
                                                            <FormControl className={ classes.formControl }>
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
                                                        <Grid item={ true } sx={ 6 }>
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
                                                                    <Typography variant={ 'subtitle1' }>
                                                                        Concurrent = runs each test on a different free
                                                                        device to get faster results<br/>
                                                                        Simultaneously = runs every test on every device
                                                                        to get a better accuracy
                                                                    </Typography>
                                                                </Grid>
                                                            </Grid>
                                                        </Grid>
                                                    </Grid>
                                                ) }
                                                { activeStep === 1 && (
                                                    <Grid container={ true } justifyContent="center" spacing={ 2 }
                                                          alignItems={ 'center' } direction={ 'column' }>
                                                        { testType === TestType.Unity && (
                                                            <div>
                                                                <Grid item={ true }>
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
                                                                            <TestMethodSelection classes={ classes }
                                                                                                 onSelectionChanged={ handleFunctionSelection }/>
                                                                        </div>
                                                                    ) }
                                                                </Grid>
                                                            </div>
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
                                <div>
                                    <Button
                                        disabled={ activeStep === 0 }
                                        onClick={ handleBack }
                                        className={ classes.backButton }
                                    >
                                        Back
                                    </Button>
                                    <Button variant="contained" color="primary" onClick={ handleNext }>
                                        { activeStep === steps.length - 1 ? 'Create' : 'Next' }
                                    </Button>
                                </div>
                            </div>
                        ) }
                    </div>
                </div>
            </Paper>
        </div>
    );
};

export default withStyles(useStyles)(AddTestPage);
