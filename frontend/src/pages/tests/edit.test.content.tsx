import React, { ChangeEvent, FC, useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import {
    Box,
    Button,
    Divider,
    FormControl,
    FormControlLabel,
    Grid,
    Radio,
    RadioGroup,
    TextField,
    Typography,
} from '@mui/material';
import { TestExecutionType } from '../../types/test.execution.type.enum';
import { TestType } from '../../types/test.type.enum';
import IDeviceData from '../../types/device';
import { useHistory } from 'react-router-dom';
import TestMethodSelection from '../../components/testmethod-selection.component';
import IAppFunctionData from '../../types/app.function';
import ITestData from '../../types/test';
import { updateTest } from '../../services/test.service';
import DeviceSelection from '../../components/device-selection.component';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles(theme => ({
    paper: {
        maxWidth: 1200,
        margin: 'auto',
        overflow: 'hidden',
    },
    searchBar: {
        borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
    },
    searchInput: {
        fontSize: theme.typography.fontSize,
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
    contentWrapper: {
        margin: '40px 16px',
    },
}));

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

interface TestContentProps {
    test: ITestData
}

const EditTestPage: FC<TestContentProps> = (props: TestContentProps) => {

    const history = useHistory();

    const { test } = props;

    const testTypes = getTestTypes();
    const executionTypes = getExecutionTypes();
    const unityTestExecutionTypes = getUnityTestsConfig();
    const deviceTypes = getDeviceOption();

    const [testName, setTestName] = useState<string>(test.Name);
    const [deviceType, setDeviceType] = useState<number>(test.TestConfig.AllDevices ? 0 : 1);
    const [executionType, setExecutionType] = useState<number>(test.TestConfig.ExecutionType);
    const [unityTestExecution, setUnityTestExecution] = useState<number>(test.TestConfig.Unity?.RunAllTests ? 0 : 1);
    const [selectedDevices, setSelectedDevices] = useState<IDeviceData[]>(test.TestConfig.Devices.map(value => value.Device));
    const [unityTestFunctions, setUnityTestFunctions] = useState<IAppFunctionData[]>(test.TestConfig.Unity?.UnityTestFunctions);

    const onTestNameChanged = (event: ChangeEvent<{ name?: string; value: string }>): void => {
        setTestName(event.target.value);
    };

    const onExecutionTypeChange = (event: ChangeEvent<{ name?: string; value: TestExecutionType }>): void => {
        const type = (+event.target.value as TestExecutionType);
        setExecutionType(type);
    };

    const onUnityTestExecutionChanged = (event: ChangeEvent<{ name?: string; value: string }>): void => {
        setUnityTestExecution(+event.target.value);
    };

    const onDeviceExecutionChanged = (event: ChangeEvent<{ name?: string; value: string }>): void => {
        setDeviceType(+event.target.value);
    };

    const onDeviceSelectionChanged = (devices: IDeviceData[]): void => {
        setSelectedDevices(devices);
    };

    const updateTestData = (): void => {
        test.Name = testName;
        test.TestConfig.AllDevices = deviceType == 0;
        test.TestConfig.ExecutionType = executionType;
        if (test.TestConfig.Unity !== null) {
            test.TestConfig.Unity.RunAllTests = unityTestExecution == 0;
        }
        updateTest(test.ID as number, test).then(response => {
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

    useEffect(() => {
        setUnityTestExecution(test.TestConfig.Unity?.RunAllTests ? 0 : 1);
        setDeviceType(test.TestConfig.AllDevices ? 0 : 1);
    }, [test]);

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
                                Edit Test
                            </Typography>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            <Button variant="contained" color="primary" size="small" onClick={ updateTestData }>
                                Save
                            </Button>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={ { p: 2, m: 2 } }>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h6' }>Test Details</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                Name:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <FormControl>
                                    <TextField required={ true } id="test-name" label="Name"
                                        value={ testName }
                                        onChange={ onTestNameChanged }/>
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
                                    value={ executionType.toString() }
                                    onChange={ onExecutionTypeChange }
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
                                    onChange={ onDeviceExecutionChanged }
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
                                            <DeviceSelection
                                                selectedDevices={ test.TestConfig.Devices.map(value => value.Device) }
                                                onSelectionChanged={ onDeviceSelectionChanged }/>
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
                                            value={ unityTestExecution.toString() }
                                            onChange={ onUnityTestExecutionChanged }
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
                                                <TestMethodSelection onSelectionChanged={ handleFunctionSelection }/>
                                            </div>
                                        ) }
                                    </Grid>
                                </Grid>
                            </div>) }
                    </Grid>
                </Grid>
            </Box>
        </Paper>
    );
};

export default EditTestPage;