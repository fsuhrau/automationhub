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
import DeviceSelection from "../../components/device-selection.component";
import IDeviceData from "../../types/device";
import { useHistory } from 'react-router-dom';
import ICreateTestData from '../../types/request.create.test';
import TestDataService from '../../services/test.service';

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        paper: {
            maxWidth: 936,
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
        input: {
            display: 'none',
        },
    }),
);
const StringIsNumber = (value): boolean => isNaN(Number(value)) === false;

function ToArray(en): Array {
    return Object.keys(en).filter(StringIsNumber).map(key => ({id: key, name: en[ key ]}));
}

function getExecutionTypes(): Array {
    return ToArray(TestExecutionType);
}

function getTestTypes(): Array {
    return ToArray(TestType);
}

function getSteps(): Array {
    return ['Select Test Type', 'Test Configuration', 'Device Selection'];
}

function getUnityTestsConfig(): Array {
    return [{id: 0, name: 'Run all Tests'}, {id: 1, name: 'Run only Selected Tests'}];
}

function getDeviceOption(): Array {
    return [{id: 0, name: 'All Devices'}, {id: 1, name: 'Selected Devices Only'}];
}

const AddTestPage: FC = (props) => {
    const classes = useStyles();
    const history = useHistory();
    const [activeStep, setActiveStep] = React.useState(0);
    const steps = getSteps();

    const testTypes = getTestTypes();
    const [testType, setTestType] = React.useState<TestType>(TestType.Unity);
    const handleTestTypeChange = (event: React.ChangeEvent<{ name?: string; value: TestType }>): void => {
        setTestType(event.target.value);
    };

    const executionTypes = getExecutionTypes();
    const [executionType, setExecutionType] = React.useState<TestExecutionType>(TestExecutionType.Parallel);
    const handleExecutionTypeChange = (event: React.ChangeEvent<{ name?: string; value: TestExecutionType }>): void => {
        setExecutionType(event.target.value);
    };

    const [testName, setTestName] = React.useState('');
    const handleTestNameChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        setTestName(event.target.value);
    };

    const unityTestExecutionTypes = getUnityTestsConfig();
    const [unityTestExecution, setUnityTestExecution] = React.useState<number>(0);
    const handleUnityTestExecutionChange = (event: React.ChangeEvent<{ name?: string; value: number }>): void => {
        setUnityTestExecution(event.target.value);
    };

    const deviceTypes = getDeviceOption();
    const [deviceType, setDeviceType] = React.useState<number>(0);
    const handleDeviceTypeChange = (event: React.ChangeEvent<{ name?: string; value: number }>): void => {
        setDeviceType(event.target.value);
    };

    const [selectedDevices, setSelectedDevices] = React.useState<IDeviceData[]>([]);
    const handleDeviceSelectionChanged = (devices: IDeviceData[]): void => {
        console.log(devices);
        setSelectedDevices(devices);
    };

    const handleNext = (): void => {
        if (activeStep === 2) {
            createTest();
        }
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
    };

    const handleBack = (): void => {
        setActiveStep((prevActiveStep) => prevActiveStep - 1);
    };

    const handleReset = (): void => {
        setActiveStep(0);
    };

    const [state, setState] = React.useState<{ age: string | number; name: string }>({
        age: '',
        name: 'hai',
    });
    const handleChange = (event: React.ChangeEvent<{ name?: string; value: unknown }>): void => {
        const name = event.target.name as keyof typeof state;
        setState({
            ...state,
            [ name ]: event.target.value,
        });
    };

    const createTest = (): void => {
        var deviceIds: number[];
        deviceIds = selectedDevices.map(value => value.ID)

        var requestData: ICreateTestData;
        requestData = {
            Name: testName,
            TestType: testType,
            UnityAllTests: unityTestExecution === 0,
            UnitySelectedTests: [],
            AllDevices: deviceType === 0,
            SelectedDevices: deviceIds,
        };

        TestDataService.create(requestData).then(response => {
            console.log(response.data);
            history.push("/tests");
        }).catch(ex => {
            console.log(ex);
        });
    }

    return (
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
                    { activeStep === steps.length  ? (
                        <div>
                            <Typography className={ classes.instructions }>Test is beeing created wait a moment and you get redirected</Typography>
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
                                                                    onChange={ handleTestTypeChange }
                                                                    inputProps={ {
                                                                        name: 'Test Type',
                                                                        id: 'test-types',
                                                                    } }>
                                                                <option aria-label="None" value=""/>
                                                                { testTypes.map((value) => (
                                                                    <option value={ value.id }>{ value.name }</option>
                                                                )) }
                                                            </Select>
                                                        </FormControl>
                                                    </Grid>
                                                    <Grid item={ true } sx={ 6 }>
                                                        <Grid container={ true } spacing={ 2 } alignItems={ 'center' }
                                                              direction={ 'row' }>
                                                            <Grid item={ true }>
                                                                <RadioGroup
                                                                    name="execution-type-selection"
                                                                    aria-label="spacing"
                                                                    value={ executionType }
                                                                    onChange={ handleExecutionTypeChange }
                                                                    row={ true }
                                                                >
                                                                    { executionTypes.map((value) => (
                                                                        <FormControlLabel
                                                                            key={ value.id }
                                                                            value={ value.id }
                                                                            control={ <Radio/> }
                                                                            label={ value.name }
                                                                        />
                                                                    )) }
                                                                </RadioGroup>
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
                                                                    name="unit-test-execution-selection"
                                                                    aria-label="spacing"
                                                                    value={ unityTestExecution }
                                                                    onChange={ handleUnityTestExecutionChange }
                                                                    row={ true }
                                                                >
                                                                    { unityTestExecutionTypes.map((value) => (
                                                                        <FormControlLabel
                                                                            key={ value.id }
                                                                            value={ value.id }
                                                                            control={ <Radio/> }
                                                                            label={ value.name }
                                                                        />
                                                                    )) }
                                                                </RadioGroup>
                                                            </Grid>
                                                            <Grid item={ true }>
                                                                { unityTestExecution === 1 && (
                                                                    <Grid container={ true } justifyContent="center"
                                                                          spacing={ 2 } alignItems={ 'center' }
                                                                          direction={ 'column' }>
                                                                        <Grid item={ true }>
                                                                            <input
                                                                                accept="image/*"
                                                                                className={ classes.input }
                                                                                id="app-upload"
                                                                                multiple={ true }
                                                                                type="file"
                                                                            />
                                                                            <label htmlFor="app-upload">
                                                                                <Button variant="contained"
                                                                                        color="primary"
                                                                                        component="span">
                                                                                    Upload
                                                                                </Button>
                                                                            </label>
                                                                        </Grid>
                                                                    </Grid>
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
                                                            value={ deviceType }
                                                            onChange={ handleDeviceTypeChange }
                                                            row={ true }
                                                        >
                                                            { deviceTypes.map((value) => (
                                                                <FormControlLabel
                                                                    key={ value.id }
                                                                    value={ value.id }
                                                                    control={ <Radio/> }
                                                                    label={ value.name }
                                                                />
                                                            )) }
                                                        </RadioGroup>
                                                    </Grid>
                                                </Grid>
                                            ) }
                                            { activeStep === 2 && deviceType === 0 && (
                                                <Grid container={ true } justifyContent="center" spacing={ 2 }
                                                      alignItems={ 'center' } direction={ 'column' }>
                                                    <Grid item={ true }>
                                                        <Typography variant={"h6"}>
                                                            Select Devices
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={ true }>
                                                        <DeviceSelection selectedDevices={selectedDevices} onSelectionChanged={handleDeviceSelectionChanged} />
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
    );
};

export default withStyles(useStyles)(AddTestPage);
