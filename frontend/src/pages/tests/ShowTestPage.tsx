import React from 'react';
import {Alert, Box, Button, Typography} from '@mui/material';
import {getTestExecutionName} from '../../types/test.execution.type.enum';
import {getTestTypeName, TestType} from '../../types/test.type.enum';
import ITestData from '../../types/test';
import {TitleCard} from "../../components/title.card.component";
import {useNavigate} from "react-router-dom";
import Grid from "@mui/material/Grid";
import {getUnityTestCategoryName, UnityTestCategory} from "../../types/unity.test.category.type.enum";

type KeyValue = { id: number, name: string };

function getUnityTestsConfig(): Array<KeyValue> {
    return [{id: 0, name: 'Run all Tests'}, {id: 1, name: 'Run only Selected Tests'}];
}

function getDeviceOption(): Array<KeyValue> {
    return [{id: 0, name: 'All Devices'}, {id: 1, name: 'Selected Devices Only'}];
}

interface TestContentProps {
    test: ITestData
}

const ShowTestPage: React.FC<TestContentProps> = (props) => {

    const {test} = props;

    const unityTestConfig = getUnityTestsConfig();
    const deviceConfig = getDeviceOption();

    const navigate = useNavigate();

    const getDeviceConfigName = (b: boolean): string => {
        const id = b ? 0 : 1;
        const item = deviceConfig.find(i => i.id === id);
        return item === undefined ? "" : item.name;
    };

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard titleElement={
                <Box sx={{display: 'flex', justifyContent: 'space-between', width: '100%'}}>
                    <Typography component="h2" variant="h6">
                        {test.Name}
                    </Typography>
                    <Button variant="contained" color="primary" size="small"
                            onClick={() => navigate(`edit`)}>Edit</Button>
                </Box>}
            >
                <TitleCard title={'Test Data'}>
                    <Grid container={true} spacing={2}>
                        <Grid size={{xs: 12, md: 2}}>
                            Type:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {getTestTypeName(test.TestConfig.Type)}
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Execution:
                        </Grid>
                        <Grid container={true} size={{xs: 12, md: 10}} spacing={2}>
                            <Grid size={{xs: 12, md: 12}}>
                                {getTestExecutionName(test.TestConfig.ExecutionType)}
                            </Grid>
                            <Grid size={{xs: 12, md: 12}}>
                                <Alert severity="info">
                                    Concurrent = runs each test on a different free
                                    device to get faster results<br/>
                                    Simultaneously = runs every test on every device
                                    to get a better accuracy</Alert>
                            </Grid>
                        </Grid>
                        <Grid size={{xs: 12, md: 2}}>
                            Devices:
                        </Grid>
                        <Grid size={{xs: 12, md: 10}}>
                            {getDeviceConfigName(test.TestConfig.AllDevices)}
                        </Grid>
                        {!test.TestConfig.AllDevices && (<>
                            <Grid size={{xs: 12, md: 2}}>
                            </Grid>
                            <Grid container={true} size={{xs: 12, md: 10}}>
                                {test.TestConfig.Devices.map((a, index) =>
                                    <Grid size={{xs: 12, md: 12}}
                                          key={`device_${a.ID}_${index}`}>- {a.Device?.DeviceIdentifier}({a.Device && (a.Device.Alias.length > 0 ? a.Device.Alias : a.Device.Name)})</Grid>,
                                )}
                            </Grid>
                        </>)}
                    </Grid>
                </TitleCard>
                <TitleCard title={'Config'}>
                    {test.TestConfig.Type === TestType.Unity && (
                        <Grid container={true} spacing={2}>
                            <Grid size={{xs: 12, md: 2}}>
                                Run:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {getUnityTestCategoryName(test.TestConfig.Unity?.UnityTestCategoryType as UnityTestCategory)}
                            </Grid>

                            {test.TestConfig.Unity?.Categories !== "" && <>
                                <Grid size={{xs: 12, md: 2}}>
                                    Categories
                                </Grid>
                                <Grid size={{xs: 12, md: 10}}>
                                    {test.TestConfig.Unity?.Categories}
                                </Grid>
                            </>
                            }
                            {test.TestConfig.Unity?.UnityTestCategoryType === UnityTestCategory.RunSelectedTestsOnly && test.TestConfig.Unity.UnityTestFunctions.length > 0 && (
                                <>
                                    <Grid size={{xs: 12, md: 2}}>
                                        Functions:
                                    </Grid>
                                    <Grid size={{xs: 12, md: 10}}>
                                        Functions:<br/>
                                        {test.TestConfig.Unity.UnityTestFunctions.map((a, index) =>
                                            <div
                                                key={`test_function_${a.ID}_${index}`}>- {a.Class}/{a.Method}<br/>
                                            </div>,
                                        )}
                                    </Grid>
                                </>)}
                        </Grid>)}
                </TitleCard>
            </TitleCard>
        </Box>
    );
};

export default ShowTestPage;
