import React, { FC, useEffect } from 'react';
import Paper from '@material-ui/core/Paper';
import { createStyles, Theme, WithStyles, withStyles } from '@material-ui/core/styles';
import {
    Box, Button,
    Divider, FormControl,
    FormControlLabel,
    Grid,
    makeStyles,
    MenuItem,
    Radio,
    RadioGroup, TextField,
    Typography,
} from '@material-ui/core';
import { TestExecutionType } from '../../types/test.execution.type.enum';
import { TestType } from '../../types/test.type.enum';
import IDeviceData from '../../types/device';
import { useHistory } from 'react-router-dom';
import TestMethodSelection from '../../components/testmethod-selection.component';
import IAppFunctionData from '../../types/app.function';
import ITestData from '../../types/test';
import { updateTest } from '../../services/test.service';
import styles = module;
import DeviceSelection from '../../components/device-selection.component';

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
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: en[ key ] }));
}

function getExecutionTypes(): Array<Object> {
    return ToArray(TestExecutionType);
}

function getTestTypes(): Array<Object> {
    return ToArray(TestType);
}

function getUnityTestsConfig(): Array<Object> {
    return [{ id: 0, name: 'Run all Tests' }, { id: 1, name: 'Run only Selected Tests' }];
}

function getDeviceOption(): Array<Object> {
    return [{ id: 0, name: 'All Devices' }, { id: 1, name: 'Selected Devices Only' }];
}

interface TestContentProps extends WithStyles<typeof styles> {
    test: ITestData
}

const EditTestPage: FC<TestContentProps> = (props) => {
    const { test, classes } = props;
    const history = useHistory();

    const testTypes = getTestTypes();
    const [testType, setTestType] = React.useState<TestType>(TestType.Unity);
    const handleTestTypeChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        const type = (+event.target.value as TestType);
        setTestType(type);
    };

    const executionTypes = getExecutionTypes();
    const handleExecutionTypeChange = (event: React.ChangeEvent<{ name?: string; value: TestExecutionType }>): void => {
        const type = (+event.target.value as TestExecutionType);
        test.TestConfig.ExecutionType = type;
    };

    const unityTestExecutionTypes = getUnityTestsConfig();
    const [unityTestExecution, setUnityTestExecution] = React.useState<number>(0);
    const handleUnityTestExecutionChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        setUnityTestExecution(+event.target.value);
        test.TestConfig.Unity.RunAllTests = (event.target.value == 0);
    };


    const deviceTypes = getDeviceOption();
    const [deviceType, setDeviceType] = React.useState<number>(0);
    const handleDeviceTypeChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        setDeviceType(+event.target.value);
        test.TestConfig.AllDevices = (event.target.value == 0);
    };

    const [selectedDevices, setSelectedDevices] = React.useState<IDeviceData[]>([]);
    const handleDeviceSelectionChanged = (devices: IDeviceData[]): void => {
        setSelectedDevices(devices);
    };

    const [unityTestFunctions, setUnityTestFunctions] = React.useState<IAppFunctionData[]>([]);

    const updateTestData = (): void => {
        updateTest(test.ID, test).then(response => {
            console.log(response.data);
            history.push('/web/tests');
        }).catch(ex => {
            console.log(ex);
        });
    };

    const handleFunctionSelection = (funcs: IAppFunctionData[]): void => {
        setUnityTestFunctions(funcs);
    };

    const getTestTypeName = (type: TestType): string => {
        const item = testTypes.find(i => i.id == type);
        return item.name;
    };

    const getTestExecutionName = (type: TestExecutionType): string => {
        const item = executionTypes.find(i => i.id == type);
        return item.name;
    };

    const handleTestNameChange = (event: React.ChangeEvent<{ name?: string; value: string }>): void => {
        test.Name = event.target.value;
    };

    useEffect(() => {
        setUnityTestExecution(test.TestConfig.Unity?.RunAllTests ? 0 : 1);
        setDeviceType(test.TestConfig.AllDevices ? 0 : 1);
    }, []);

    return (
        <div>
            <Grid container={ true } sx={ { p: 2, m: 2 } }>
                <Grid item={ true } xs={ 12 }>
                    <Box component={ Paper } sx={ { p: 2, m: 2 } }>
                        <Typography variant={ 'h6' }>Test Details</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                Name:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <FormControl className={ classes.formControl }>
                                    <TextField required={ true } id="test-name" label="Name"
                                        value={ test.Name }
                                        onChange={ handleTestNameChange }/>
                                </FormControl>
                            </Grid>
                        </Grid>
                        <br/>
                        <Typography variant={ 'h6' }>Test Configuration</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                Type:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { getTestTypeName(test.TestConfig.Type) }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Execution:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <RadioGroup
                                    name="execution-type-selection"
                                    aria-label="spacing"
                                    value={ test.TestConfig.ExecutionType.toString() }
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
                                <br/>
                                <Typography variant={ 'subtitle1' }>
                                    Concurrent = runs each test on a different free
                                    device to get faster results<br/>
                                    Simultaneously = runs every test on every device
                                    to get a better accuracy
                                </Typography>
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Devices:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
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
                                { deviceType === 1 && (
                                    <Grid container={ true } justifyContent="center" spacing={ 2 }
                                        alignItems={ 'center' } direction={ 'column' }>
                                        <Grid item={ true }>
                                            <Typography variant={ 'h6' }>
                                                Select Devices
                                            </Typography>
                                        </Grid>
                                        <Grid item={ true }>
                                            <DeviceSelection selectedDevices={ test.TestConfig.Devices }
                                                onSelectionChanged={ handleDeviceSelectionChanged }/>
                                        </Grid>
                                    </Grid>
                                ) }
                            </Grid>
                        </Grid>

                        { test.TestConfig.Type === TestType.Unity && (
                            <div>
                                <br/>
                                <Typography variant={ 'h6' }>Unity Config</Typography>
                                <Divider/>
                                <br/>
                                <Grid container={ true }>
                                    <Grid item={ true } xs={ 2 }>
                                        Execute Tests:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        <RadioGroup
                                            name="unity-test-execution-selection"
                                            aria-label="spacing"
                                            value={ test.TestConfig.Unity?.RunAllTests ? 'ß' : '1' }
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

                                        { unityTestExecution === 1 && (
                                            <div>
                                                <TestMethodSelection classes={ classes }
                                                    onSelectionChanged={ handleFunctionSelection }/>
                                            </div>
                                        ) }

                                        { test.TestConfig.Unity?.RunAllTests === true && (<div>all</div>) }
                                        { test.TestConfig.Unity?.RunAllTests === false && (<div>
                                            { test.TestConfig.Unity.UnityTestFunctions.map((a) =>
                                                <div>- { a.Class }/{ a.Method }<br/></div>,
                                            ) }
                                        </div>) }
                                    </Grid>
                                </Grid>
                            </div>) }
                    </Box>
                </Grid>
                <Grid item={ true }>
                    <Button variant="contained" color="primary" onClick={ updateTestData }>
                        Save
                    </Button>
                </Grid>
            </Grid>
        </div>
    );
};

export default withStyles(useStyles)(EditTestPage);
