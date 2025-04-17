import React, {useEffect, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    FormControl,
    FormControlLabel,
    Radio,
    RadioGroup,
    TextField,
    Typography,
} from '@mui/material';
import {getExecutionTypes, TestExecutionType} from '../../types/test.execution.type.enum';
import {getTestTypes, TestType} from '../../types/test.type.enum';
import {useNavigate} from 'react-router-dom';
import TestMethodSelection from '../../components/testmethod-selection.component';
import ITestData from '../../types/test';
import {updateTest, UpdateTestData} from '../../services/test.service';
import DeviceSelection from '../../components/device-selection.component';
import {PlatformType} from '../../types/platform.type.enum';
import Table from "@mui/material/Table";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import TableBody from "@mui/material/TableBody";
import IconButton from "@mui/material/IconButton";
import {Delete} from "@mui/icons-material";
import IAppFunctionData from "../../types/app.function";
import {getDeviceOption, getUnityTestsConfig} from "./add.test.content";
import {getAllDevices} from "../../services/device.service";
import IDeviceData from "../../types/device";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {isEqual} from "lodash";
import {UnityTestCategory} from "../../types/unity.test.category.type.enum";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import Grid from "@mui/material/Grid";
import {TitleCard} from "../../components/title.card.component";
import {useError} from "../../ErrorProvider";

interface TestContentProps {
    test: ITestData
}

const EditTestPage: React.FC<TestContentProps> = (props: TestContentProps) => {

    const {project, projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {setError} = useError()

    const navigate = useNavigate();

    type NewTestState = {
        testName: string,
        testType: TestType,
        executionType: TestExecutionType,
        platformType: PlatformType,
        unityTestCategoryType: UnityTestCategory,
        deviceType: number,
        selectedDevices: number[],
        selectedTestFunctions: IAppFunctionData[],
        testCategories: string[],
        category: string,
    };

    const {test} = props;

    const app = project.Apps.find(a => a.ID === appId);

    const testTypes = getTestTypes();
    const executionTypes = getExecutionTypes();
    const unityTestExecutionTypes = getUnityTestsConfig();
    const deviceTypes = getDeviceOption();

    const [uiState, setUiState] = React.useState<NewTestState>({
            testType: TestType.Unity,
            executionType: test.TestConfig.ExecutionType,
            platformType: PlatformType.iOS,
            testName: test.Name,
            unityTestCategoryType: test.TestConfig.Unity === undefined || test.TestConfig.Unity === null ? UnityTestCategory.RunAllTests : test.TestConfig.Unity?.UnityTestCategoryType,
            deviceType: test.TestConfig.AllDevices ? 0 : 1,
            selectedDevices: test.TestConfig.Devices.map(value => value.DeviceID) as number[],
            selectedTestFunctions: test.TestConfig.Unity === undefined || test.TestConfig.Unity === null ? [] : test.TestConfig.Unity.UnityTestFunctions.map(value => ({
                Assembly: value.Assembly,
                Class: value.Class,
                Method: value.Method
            } as IAppFunctionData)),
            testCategories: test.TestConfig.Unity === undefined || test.TestConfig.Unity === null || test.TestConfig.Unity.Categories === '' ? [] : test.TestConfig.Unity.Categories.split(','),
            category: '',
        }
    )

    const addCategory = () => {
        if (uiState.category != '') {
            setUiState(prevState => ({
                ...prevState,
                testCategories: [...prevState.testCategories, prevState.category],
                category: ''
            }))
        }
    };
    const removeCategory = (index: number) => {
        setUiState(prevState => ({
            ...prevState,
            testCategories: [...prevState.testCategories.slice(0, index), ...prevState.testCategories.slice(index + 1)]
        }))
    };

    const updateTestData = (): void => {
        updateTest(projectIdentifier, appId, test.ID as number, {
            Name: uiState.testName,
            Categories: uiState.testCategories.join(','),
            AllDevices: uiState.deviceType === 0,
            ExecutionType: uiState.executionType,
            UnityTestCategoryType: uiState.unityTestCategoryType,
            Devices: uiState.selectedDevices,
            TestFunctions: uiState.selectedTestFunctions,
        } as UpdateTestData).then(response => {
            navigate(`/project/${projectIdentifier}/app:${appId}/tests`);
        }).catch(ex => {
            setError(ex);
        });
    };

    const getTestTypeName = (type: TestType): string => {
        const item = testTypes.find(i => i.id === `${type}`);
        return item === undefined ? '' : item.name;
    };

    const [devices, setDevices] = useState<IDeviceData[]>([]);

    useEffect(() => {
        getAllDevices(projectIdentifier, app?.Platform).then(response => {
            setDevices(response.data);
        }).catch(ex => setError(ex))
    }, [projectIdentifier, app?.Platform])

    const onDeviceSelectionChanged = (selectedDevices: number[]) => {
        if (!isEqual(selectedDevices, uiState.selectedDevices)) {
            setUiState(prevState => ({...prevState, selectedDevices: selectedDevices}))
        }
    }

    const onTestSelectionChanged = (testSelection: IAppFunctionData[]) => {
        if (!isEqual(testSelection, uiState.selectedTestFunctions)) {
            setUiState(prevState => ({...prevState, selectedTestFunctions: testSelection}))
        }
    }

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <Typography
                component="h2" variant="h6" sx={{mb: 2}}>
                Edit Test
            </Typography>
            <Grid
                container
                spacing={2}
                columns={12}
                sx={{mb: (theme) => theme.spacing(2)}}
            >
            </Grid>
            <TitleCard title={'Test Details'}>
                <Grid container={true}>
                    <Grid size={{xs: 12, md: 2}}>
                        Name:
                    </Grid>
                    <Grid size={{xs: 12, md: 10}}>
                        <TextField required={true} id="test-name" placeholder="Name"
                                   value={uiState.testName}
                                   onChange={event => setUiState(prevState => ({
                                       ...prevState,
                                       testName: event.target.value
                                   }))}/>
                    </Grid>
                </Grid>
            </TitleCard>
            <TitleCard title={'Test Configuration'}>
                <Grid container={true}>
                    <Grid size={{xs: 12, md: 2}}>
                        Type:
                    </Grid>
                    <Grid size={{xs: 12, md: 10}}>
                        {getTestTypeName(test.TestConfig.Type)}
                    </Grid>

                    <Grid size={{xs: 12, md: 2}}>
                        Execution:
                    </Grid>
                    <Grid size={{xs: 12, md: 10}}>
                        <RadioGroup
                            name="execution-type-selection"
                            aria-label="spacing"
                            value={uiState.executionType.toString()}
                            onChange={event => setUiState(prevState => ({
                                ...prevState,
                                executionType: +event.target.value
                            }))}
                            row={true}
                        >
                            {executionTypes.map((value) => (
                                <FormControlLabel
                                    key={'exec_' + value.id}
                                    value={value.id.toString()}
                                    control={<Radio/>}
                                    label={value.name}
                                />
                            ))}
                        </RadioGroup>
                    </Grid>

                    <Grid size={{xs: 12, md: 2}}>
                    </Grid>
                    <Grid size={{xs: 12, md: 10}}>
                        <Alert severity="info">
                            Concurrent = runs each test on a different free
                            device to get faster results<br/>
                            Simultaneously = runs every test on every device
                            to get a better accuracy</Alert>
                    </Grid>

                    <Grid size={{xs: 12, md: 2}}>
                        Devices:
                    </Grid>
                    <Grid size={{xs: 12, md: 10}}>
                        <RadioGroup
                            name="device-selection"
                            aria-label="spacing"
                            value={uiState.deviceType.toString()}
                            onChange={event => setUiState(prevState => ({
                                ...prevState,
                                deviceType: +event.target.value
                            }))}
                            row={true}
                        >
                            {deviceTypes.map((value) => (
                                <FormControlLabel
                                    key={'device_' + value.id}
                                    value={value.id.toString()}
                                    control={<Radio/>}
                                    label={value.name}
                                />
                            ))}
                        </RadioGroup>
                        {uiState.deviceType === 1 && (
                            devices !== undefined && devices.length > 0 ?
                                <DeviceSelection
                                    devices={devices}
                                    selectedDevices={uiState.selectedDevices}
                                    onSelectionChanged={onDeviceSelectionChanged}/>
                                : <Alert severity="error">No devices connected</Alert>
                        )}
                    </Grid>
                </Grid>
            </TitleCard>
            {test.TestConfig.Type === TestType.Unity && (
                <TitleCard title={'Unity Test Config'}>
                    <Grid container={true}>

                        <Grid size={{xs: 12, md: 2}}>
                            Execute Tests:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            <RadioGroup
                                name="unity-test-execution-selection"
                                aria-label="spacing"
                                value={(+uiState.unityTestCategoryType).toString()}
                                onChange={event => setUiState(prevState => ({
                                    ...prevState,
                                    unityTestCategoryType: +event.target.value
                                }))}
                                row={true}
                            >
                                {unityTestExecutionTypes.map((value) => (
                                    <FormControlLabel
                                        key={'unityt_' + value.id}
                                        value={(+value.id).toString()}
                                        control={<Radio/>}
                                        label={value.name}
                                    />
                                ))}
                            </RadioGroup>
                        </Grid>

                        <Grid size={{xs: 12, md: 2}}>
                        </Grid>

                        {uiState.unityTestCategoryType === UnityTestCategory.RunAllOfCategory && (
                            <Grid size={{xs: 12, md: 10}} container={true}>
                                <Grid size={{xs: 12, md: 12}}>
                                    Categories
                                </Grid>
                                <Grid size={{xs: 12, md: 12}}>
                                    <Table size="small" aria-label="a dense table" width={600}>
                                        <TableHead>
                                            <TableRow>
                                                <TableCell>Category</TableCell>
                                                <TableCell></TableCell>
                                            </TableRow>
                                        </TableHead>
                                        <TableBody>
                                            {uiState.testCategories.map((element, index) => <TableRow
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
                                                    <TextField required={true}
                                                               id="test-name"
                                                               placeholder="Category Name"
                                                               value={uiState.category}
                                                               fullWidth={true}
                                                               onChange={event => setUiState(prevState => ({
                                                                   ...prevState,
                                                                   category: event.target.value
                                                               }))}
                                                    />
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

                        {uiState.unityTestCategoryType === UnityTestCategory.RunSelectedTestsOnly && (
                            <Grid size={{xs: 12, md: 10}} container={true}>
                                <Grid size={{xs: 12, md: 12}}>
                                    Selected Tests:
                                </Grid>
                                <Grid size={{xs: 12, md: 12}}>
                                    <TestMethodSelection selectedTestFunctions={uiState.selectedTestFunctions} onSelectionChanged={onTestSelectionChanged}/>
                                </Grid>
                            </Grid>
                        )}
                    </Grid>
                </TitleCard>
            )}
            <Box sx={{display: 'flex', justifyContent: 'space-between', width: '100%'}}>
                <Typography component="h2" variant="h6">
                </Typography>
                <Button variant="contained" color="primary" size="small" onClick={updateTestData}>
                    Save
                </Button>
            </Box>
        </Box>
    );
};

export default EditTestPage;