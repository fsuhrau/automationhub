import React, { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import {
    Alert,
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
import { getExecutionTypes, TestExecutionType } from '../../types/test.execution.type.enum';
import { getTestTypes, TestType } from '../../types/test.type.enum';
import { useHistory } from 'react-router-dom';
import TestMethodSelection from '../../components/testmethod-selection.component';
import ITestData from '../../types/test';
import { updateTest } from '../../services/test.service';
import DeviceSelection from '../../components/device-selection.component';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { getPlatformTypes, PlatformType } from '../../types/platform.type.enum';
import Table from "@mui/material/Table";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import TableBody from "@mui/material/TableBody";
import IconButton from "@mui/material/IconButton";
import { Delete } from "@mui/icons-material";
import IAppFunctionData from "../../types/app.function";
import { getDeviceOption, getUnityTestsConfig } from "./add.test.content";
import { getAllDevices } from "../../services/device.service";
import IDeviceData from "../../types/device";

// Warn if overriding existing method
if(Array.prototype.equals)
    console.warn("Overriding existing Array.prototype.equals. Possible causes: New API defines the method, there's a framework conflict or you've got double inclusions in your code.");
// attach the .equals method to Array's prototype to call it on any array
Array.prototype.equals = function (array) {
    // if the other array is a falsy value, return
    if (!array)
        return false;

    // compare lengths - can save a lot of time
    if (this.length != array.length)
        return false;

    for (var i = 0, l=this.length; i < l; i++) {
        // Check if we have nested arrays
        if (this[i] instanceof Array && array[i] instanceof Array) {
            // recurse into the nested arrays
            if (!this[i].equals(array[i]))
                return false;
        }
        else if (this[i] != array[i]) {
            // Warning - two different object instances will never be equal: {x:20} != {x:20}
            return false;
        }
    }
    return true;
}
// Hide method from for-in loops
Object.defineProperty(Array.prototype, "equals", {enumerable: false});

interface TestContentProps {
    test: ITestData
}

const EditTestPage: React.FC<TestContentProps> = (props: TestContentProps) => {

    const history = useHistory();

    type NewTestState = {
        testName: string,
        testType: TestType,
        executionType: TestExecutionType,
        platformType: PlatformType,
        unityTestExecutionType: number,
        deviceType: number,
        selectedDevices: number[],
        selectedTestFunctions: IAppFunctionData[],
        testCategories: string[],
        category: string,
    };

    const {test} = props;

    const testTypes = getTestTypes();
    const executionTypes = getExecutionTypes();
    const platformTypes = getPlatformTypes();
    const unityTestExecutionTypes = getUnityTestsConfig();
    const deviceTypes = getDeviceOption();

    const addCategory = () => {
        if (state.category != '') {
            setState(prevState => ({...prevState, testCategories: [...prevState.testCategories, prevState.category], category: ''}))
        }
    };
    const removeCategory = (index: number) => {
        setState(prevState => ({...prevState, testCategories: [...prevState.testCategories.slice(0, index), ...prevState.testCategories.slice(index+1)]}))
    };

    const [state, setState] = React.useState<NewTestState>({
            testType: TestType.Unity,
            executionType: test.TestConfig.ExecutionType,
            platformType: PlatformType.iOS,
            testName: test.Name,
            unityTestExecutionType: test.TestConfig.Unity?.RunAllTests ? 0 : 1,
            deviceType: test.TestConfig.AllDevices ? 0 : 1,
            selectedDevices: test.TestConfig.Devices.map(value => value.ID) as number[],
            selectedTestFunctions: test.TestConfig.Unity === undefined || test.TestConfig.Unity === null ? [] : test.TestConfig.Unity.UnityTestFunctions.map(value => ({Assembly: value.Assembly, Class: value.Class, Method: value.Method} as IAppFunctionData)),
            testCategories: test.TestConfig.Unity === undefined || test.TestConfig.Unity === null ? [] : test.TestConfig.Unity.Categories.split(','),
            category: '',
        }
    )
    const updateTestData = (): void => {
        test.Name = state.testName;
        test.TestConfig.AllDevices = state.deviceType === 0;
        test.TestConfig.ExecutionType = state.executionType;
        if (test.TestConfig.Unity !== null && test.TestConfig.Unity !== undefined) {
            test.TestConfig.Unity.RunAllTests = state.unityTestExecutionType == 0;
            // test.TestConfig.Devices = state.selectedDevices;
        }
        updateTest(test.ID as number, test).then(response => {
            console.log(response.data);
            history.push('/web/tests');
        }).catch(ex => {
            console.log(ex);
        });
    };

    const getTestTypeName = (type: TestType): string => {
        const item = testTypes.find(i => i.id == type);
        return item.name;
    };

    const [devices, setDevices] = useState<IDeviceData[]>([]);

    useEffect(() => {
        getAllDevices().then(response => {
            setDevices(response.data);
        })
    }, [])

    const onDeviceSelectionChanged = (selectedDevices: number[]) => {
        if (selectedDevices.equals(state.selectedDevices) === false) {
            setState(prevState => ({...prevState, selectedDevices: selectedDevices}))
        }
    }

    const onTestSelectionChanged = (testSelection: IAppFunctionData[]) => {
        if (testSelection.equals(state.selectedTestFunctions) === false) {
            setState(prevState => ({...prevState, selectedTestFunctions: testSelection}))
        }
    }

    return (
        <Paper sx={ {maxWidth: 1200, margin: 'auto', overflow: 'hidden'} }>
            <AppBar
                position="static"
                color="default"
                elevation={ 0 }
                sx={ {borderBottom: '1px solid rgba(0, 0, 0, 0.12)'} }
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
            <Box sx={ {p: 2, m: 2} }>
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
                                               value={ state.testName }
                                               onChange={ event => setState(prevState => ({
                                                   ...prevState,
                                                   testName: event.target.value
                                               })) }/>
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
                                    value={ state.executionType.toString() }
                                    onChange={ event => setState(prevState => ({
                                        ...prevState,
                                        executionType: +event.target.value
                                    })) }
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
                                <Alert severity="info">
                                    Concurrent = runs each test on a different free
                                    device to get faster results<br/>
                                    Simultaneously = runs every test on every device
                                    to get a better accuracy</Alert>
                                <br/>
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Devices:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <RadioGroup
                                    name="device-selection"
                                    aria-label="spacing"
                                    value={ state.deviceType.toString() }
                                    onChange={ event => setState(prevState => ({
                                        ...prevState,
                                        deviceType: +event.target.value
                                    })) }
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
                                { state.deviceType === 1 && (
                                    <Grid container={ true } justifyContent="center" spacing={ 2 }
                                          alignItems={ 'center' } direction={ 'column' }>
                                        <Grid item={ true }>
                                            <Typography variant={ 'h6' }>
                                                Select Devices
                                            </Typography>
                                        </Grid>
                                        <Grid item={ true }>
                                            <DeviceSelection
                                                devices={devices}
                                                selectedDevices={ test.TestConfig.Devices.map(value => value.DeviceID) }
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
                                            value={ state.unityTestExecutionType.toString() }
                                            onChange={ event => setState(prevState => ({...prevState, unityTestExecutionType: +event.target.value})) }
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
                                    { state.unityTestExecutionType === 2 && (
                                        <Grid item={true} xs={12} container={true} >
                                            <Grid item={ true } xs={ 2 }>
                                                Selected Tests:
                                            </Grid>
                                            <Grid item={ true } xs={ 10 }>
                                                <TestMethodSelection onSelectionChanged={ onTestSelectionChanged }/>
                                            </Grid>
                                        </Grid>
                                    ) }
                                    { state.unityTestExecutionType === 1 && (
                                        <Grid item={true} xs={12} container={true} >
                                            <Grid item={ true } xs={ 2 }>
                                                Categories
                                            </Grid>
                                            <Grid item={ true } xs={ 10 }>
                                                <Table size="small" aria-label="a dense table" width={ 600 }>
                                                    <TableHead>
                                                        <TableRow>
                                                            <TableCell>Category</TableCell>
                                                            <TableCell></TableCell>
                                                        </TableRow>
                                                    </TableHead>
                                                    <TableBody>
                                                        { state.testCategories.map((element, index) => <TableRow
                                                            key={ `categories_${ index }` }>
                                                            <TableCell>{ element }</TableCell>
                                                            <TableCell>
                                                                <IconButton edge="end"
                                                                            aria-label="delete"
                                                                            onClick={ () => removeCategory(index) }>
                                                                    <Delete/>
                                                                </IconButton>
                                                            </TableCell>
                                                        </TableRow>) }

                                                        <TableRow>
                                                            <TableCell>
                                                                <FormControl>
                                                                    <TextField required={ true }
                                                                               id="test-name"
                                                                               label="Category Name"
                                                                               value={ state.category }
                                                                               fullWidth={ true }
                                                                               onChange={ event => setState(prevState => ({
                                                                                   ...prevState,
                                                                                   category: event.target.value
                                                                               })) }
                                                                    />
                                                                </FormControl>
                                                            </TableCell>
                                                            <TableCell>
                                                                <Button variant={ 'outlined' }
                                                                        onClick={ addCategory }>Add</Button>
                                                            </TableCell>
                                                        </TableRow>
                                                    </TableBody>
                                                </Table>
                                            </Grid>
                                        </Grid>
                                    ) }
                                </Grid>
                            </div>) }
                    </Grid>
                </Grid>
            </Box>
        </Paper>
    );
};

export default EditTestPage;