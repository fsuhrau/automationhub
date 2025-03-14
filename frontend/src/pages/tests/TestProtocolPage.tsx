import React, {FunctionComponent, ReactElement, useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import ITestRunData from '../../types/test.run';
import {TestResultState} from '../../types/test.result.state.enum';
import ITestProtocolData, {duration} from '../../types/test.protocol';
import Typography from '@mui/material/Typography';
import {AvTimer, DateRange, Deblur, PhoneAndroid, Speed} from '@mui/icons-material';
import {Button, Card, CardMedia, Divider, Popover, Tab, Tabs} from '@mui/material';
import moment from 'moment';
import IProtocolEntryData from '../../types/protocol.entry';
import TestStatusIconComponent from '../../components/test-status-icon.component';
import {useSSE} from 'react-hooks-sse';
import ProtocolLogComponent from '../../components/protocol.log.component';
import ProtocolScreensComponent from '../../components/protocol.screens.component';
import IProtocolPerformanceEntryData from '../../types/protocol.performance.entry';
import {CartesianGrid, LabelList, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";
import {TitleCard} from "../../components/title.card.component";
import {Box} from "@mui/system";
import Grid from "@mui/material/Grid2";

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

function TabPanel(props: TabPanelProps): ReactElement {
    const {children, value, index, ...other} = props;

    return (
        <Grid container={true} size={{xs: 12, md: 12}}
              role="tabpanel"
              hidden={value !== index}
              id={`simple-tabpanel-${index}`}
              aria-labelledby={`simple-tab-${index}`}
              {...other}
        >
            {value === index && (
                <>
                    {children}
                </>
            )}
        </Grid>
    );
}

function a11yProps(index: number): Map<string, string> {
    return new Map([
        ['id', `simple-tab-${index}`],
        ['aria-controls', `simple-tabpanel-${index}`],
    ]);
}

interface NewTestProtocolLogPayload {
    TestProtocolID: number,
    Entry: IProtocolEntryData,
}

interface TestProtocolPageProps {
    run: ITestRunData
    protocol: ITestProtocolData
}

const CustomLineLabel: FunctionComponent<any> = (props: any) => {
    const {x, y, stroke, value, unit} = props;
    return (
        <text x={x} y={y} dy={-4} fill={stroke} fontSize={10} textAnchor="middle">
            {value.toFixed(2)}{unit !== undefined ? unit : ''}
        </text>
    );
};

interface PerformanceNoteLabelProps {
    value: number,
    isHigherBetter: boolean,
}

const PerformanceNoteLabel: React.FC<PerformanceNoteLabelProps> = (props: PerformanceNoteLabelProps) => {
    const {value, isHigherBetter} = props;
    let color = 'red';
    if (isHigherBetter) {
        if (value > 0) {
            color = 'green';
        }
    } else {
        if (value < 0) {
            color = 'green';
        }
    }
    return (
        <span style={{
            color: color,
            fontSize: '0.8em',
            verticalAlign: "top"
        }}>{value >= 0 ? '+' : ''}{value.toFixed(2)}</span>
    );
};

const TestProtocolPage: React.FC<TestProtocolPageProps> = (props) => {

    const {protocol} = props;

    const [anchorScreenEl, setAnchorScreenEl] = useState<HTMLButtonElement | null>(null);
    const showScreenPopup = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorScreenEl(event.currentTarget);
    };
    const hideScreenPopup = (): void => {
        setAnchorScreenEl(null);
    };

    const lastScreenOpen = Boolean(anchorScreenEl);
    const lastScreenID = lastScreenOpen ? 'simple-popover' : undefined;

    const [state, setState] = useState<{
        lastScreen: IProtocolEntryData | null,
        lastErrors: IProtocolEntryData[],
        lastStep: IProtocolEntryData | null,
        entries: IProtocolEntryData[],
        screenEntries: IProtocolEntryData[],
        performanceEntries: IProtocolPerformanceEntryData[],
        steps: number
    }>({
        lastScreen: null,
        lastErrors: [],
        lastStep: null,
        entries: [],
        screenEntries: [],
        performanceEntries: [],
        steps: 0,
    })

    const protocolEntry = useSSE<NewTestProtocolLogPayload | null>(`test_protocol_${protocol.ID}_log`, null);
    useEffect(() => {
        if (protocolEntry === null)
            return;
        setState(prevState => ({...prevState, entries: [...prevState.entries, protocolEntry.Entry]}))
    }, [protocolEntry]);

    const updateStatusEntries = (): void => {
        const length = state.entries.length;
        let numSteps = 0;
        const errors: IProtocolEntryData[] = [];
        let lastStep: IProtocolEntryData | null = null;
        let lastScreen: IProtocolEntryData | null = null;
        const screenEntries: IProtocolEntryData[] = [];
        for (let i = length - 1; i > 0; i--) {
            if (state.entries[i].Source === 'screen') {
                screenEntries.push(state.entries[i])
                if (lastScreen === null) {
                    lastScreen = state.entries[i];
                }
            }
            if (state.entries[i].Source === 'step') {
                numSteps++;
                if (lastStep === null) {
                    lastStep = state.entries[i];
                }
            }
            if (state.entries[i].Level === 'error') {
                errors.push(state.entries[i]);
            }
        }
        setState(prevState => ({
            ...prevState,
            lastScreen: lastScreen,
            lastStep: lastStep,
            lastErrors: errors,
            steps: numSteps,
            screenEntries: screenEntries
        }))
    };

    const [value, setValue] = useState(0);
    const handleChange = (event: React.ChangeEvent<{}>, newValue: number): void => {
        setValue(newValue);
    };

    useEffect(() => {
        setState(prevState => ({...prevState, entries: protocol.Entries, performanceEntries: protocol.Performance}))
    }, [protocol]);

    useEffect(() => {
        updateStatusEntries();
    }, [state.entries]);

    interface Checkpoints {
        Name: string,
        Runtime: number,
        ExecutionTime: number,
    }

    const checkpoints = state.performanceEntries.filter(p => p.Checkpoint !== "schedule").map((value1, index, array) => {
        if (index !== 0) {
            value1.ExecutionTime = value1.Runtime - array[index - 1].Runtime
        } else {
            value1.ExecutionTime = value1.Runtime
        }
        return value1
    })

    const historicalCheckpoints = protocol.TestProtocolHistory.slice(-2).map(p => {
        p.Performance = p.Performance.filter(p => p.Checkpoint !== "schedule").map((value1, index, array) => {
            if (index !== 0) {
                value1.ExecutionTime = value1.Runtime - array[index - 1].Runtime
            } else {
                value1.ExecutionTime = value1.Runtime
            }
            return value1
        })
        return p
    })

    const diffAvgFPS = protocol.AvgFPS - protocol.HistAvgFPS;
    const diffAvgMEM = protocol.AvgMEM - protocol.HistAvgMEM;
    const diffAvgCPU = protocol.AvgCPU - protocol.HistAvgCPU;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={protocol.TestName}>
                <Grid container={true} spacing={1} alignItems="center" sx={{
                    padding: 1,
                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                }}>
                    <Grid>
                        <TestStatusIconComponent
                            status={protocol.TestResult}/>
                    </Grid>
                    <Grid>
                        <DateRange sx={{display: 'block'}} color="inherit"/>
                    </Grid>
                    <Grid>
                        {moment(protocol.StartedAt).format('YYYY/MM/DD HH:mm:ss')}
                    </Grid>

                    <Grid>
                        <AvTimer sx={{display: 'block'}} color="inherit"/>
                    </Grid>

                    <Grid>
                        {duration(protocol.StartedAt, protocol.EndedAt)}
                    </Grid>

                    {protocol.Device && <Grid>
                        <PhoneAndroid sx={{display: 'block'}} color="inherit"/>
                    </Grid>}
                    {protocol.Device && <Grid>
                        {protocol.Device && (protocol.Device.Alias.length > 0 ? protocol.Device.Alias : protocol.Device.Name)}
                    </Grid>}

                    {/**/
                        <Grid>
                            <Speed sx={{display: 'block'}} color="inherit"/>
                        </Grid>
                    }
                </Grid>
            </TitleCard>


            <TitleCard title={"Test Protocol"}>
                <Paper sx={{width: '100%', overflow: 'hidden'}}>
                    <Grid container={true}>
                        <Tabs
                            value={value}
                            onChange={handleChange}
                            indicatorColor="primary"
                            textColor="inherit"
                        >
                            <Tab label="Status" {...a11yProps(0)} />
                            <Tab label="Logs" {...a11yProps(1)} />
                            <Tab label="Screenshots" {...a11yProps(2)} />
                            <Tab label="Video/Replay" {...a11yProps(3)} />
                            <Tab label="Performance" {...a11yProps(4)} />
                            <Tab label="Checkpoints" {...a11yProps(5)} />
                        </Tabs>
                    </Grid>

                    <Grid container={true}>
                        <TabPanel value={value} index={0}>
                            <Grid container={true} size={{xs: 12, md: 12}}>
                                {protocol?.TestResult == TestResultState.TestResultFailed &&
                                    <Grid container={true} sx={{padding: 1, backgroundColor: '#ff2b40'}} size={{xs: 12, md: 12}}>
                                        <Grid container={true} size={{xs: 12, md: 12}} spacing={2}>
                                            {state.lastStep &&
                                                <Grid container={true} size={12}>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography
                                                            variant={"body2"}>{moment(state.lastStep.CreatedAt).format('YYYY/MM/DD HH:mm:ss')}</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}}>
                                                        <Typography
                                                            variant={"body2"}>{state.lastStep.Message}</Typography>
                                                    </Grid>
                                                    <Grid size={12}>
                                                        <Divider />
                                                    </Grid>
                                                </Grid>
                                            }
                                            {state.lastErrors.map((lastError, index) =>
                                                (<Grid key={`last_error_${index}`} container={true} size={12}>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography
                                                            variant={"body2"}>{moment(lastError.CreatedAt).format('YYYY/MM/DD HH:mm:ss')}</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}}>
                                                        <Typography
                                                            variant={"body2"}>{lastError.Message}</Typography>
                                                    </Grid>
                                                    <Grid size={12}>
                                                        <Divider />
                                                    </Grid>
                                                </Grid>)
                                            )}
                                        </Grid>
                                        <Grid size={{xs: 12, md: 12}} container={true}
                                              justifyContent={"center"} alignItems={"center"} justifyItems={"center"}>
                                            {state.lastScreen && <div>
                                                <Button aria-describedby={'last_screen_' + state.lastScreen.ID}
                                                        variant="contained" onClick={showScreenPopup}>
                                                    Show
                                                </Button>
                                                <Popover
                                                    id={lastScreenID}
                                                    open={lastScreenOpen}
                                                    anchorEl={anchorScreenEl}
                                                    onClose={hideScreenPopup}
                                                    anchorOrigin={{
                                                        vertical: 'bottom',
                                                        horizontal: 'left',
                                                    }}
                                                >
                                                    <Card>
                                                        <CardMedia
                                                            component="img"
                                                            height="400"
                                                            image={`/api/data/${state.lastScreen.Data}`}
                                                            alt="green iguana"
                                                        />
                                                    </Card>
                                                </Popover>
                                            </div>}
                                        </Grid>
                                    </Grid>
                                }
                                <Grid size={{xs: 12, md: 12}} container={true} sx={{padding: 1}}>
                                    <Grid size={{xs: 12, md: 3}} container={true} spacing={2}>
                                        <Grid size={12}>
                                            <Typography gutterBottom={true} variant="subtitle1">
                                                Time
                                            </Typography>
                                        </Grid>
                                        <Grid>
                                            <Typography gutterBottom={true} variant="body2"
                                                        color="textSecondary">
                                                Execution
                                            </Typography>
                                            <Typography variant="h5" style={{whiteSpace: 'nowrap'}}>
                                                {duration(protocol.StartedAt, protocol.EndedAt)}
                                            </Typography>
                                        </Grid>
                                    </Grid>
                                    <Grid size={{xs: 12, md: 9}} container={true}>
                                        <Grid size={12}>
                                            <Typography gutterBottom={true} variant="subtitle1">
                                                Details
                                            </Typography>
                                        </Grid>
                                        <Grid container={true} spacing={5}>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">Steps</Typography>
                                                <Typography variant="h5">{state.steps}</Typography>
                                            </Grid>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                    FPS
                                                </Typography>
                                                <Typography variant="h5">
                                                    {protocol.AvgFPS?.toFixed(2)} <PerformanceNoteLabel
                                                    value={diffAvgFPS} isHigherBetter={true}/>
                                                </Typography>
                                            </Grid>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                    Memory
                                                </Typography>
                                                <Typography variant="h5" component="h5">
                                                    {protocol.AvgMEM?.toFixed(2)}MB <PerformanceNoteLabel
                                                    value={diffAvgMEM} isHigherBetter={false}/>
                                                </Typography>
                                            </Grid>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                    CPU
                                                </Typography>
                                                <Typography variant="h5" component="h5">
                                                    {protocol.AvgCPU?.toFixed(2)}% <PerformanceNoteLabel
                                                    value={diffAvgCPU} isHigherBetter={false}/>
                                                </Typography>
                                            </Grid>
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Grid>
                        </TabPanel>
                        <TabPanel value={value} index={1}>
                            <ProtocolLogComponent entries={state.entries}/>
                        </TabPanel>
                        <TabPanel value={value} index={2}>
                            <ProtocolScreensComponent entries={state.screenEntries}/>
                        </TabPanel>
                        <TabPanel value={value} index={3}>
                            Video (not implemented)
                        </TabPanel>
                        <TabPanel value={value} index={4}>
                            <Grid container={true} sx={{padding: 1}} spacing={1}>
                                <Grid size={{xs: 12, md: 12}}>
                                    <Typography gutterBottom={true} variant="subtitle1">
                                        FPS
                                    </Typography>
                                    <Divider/>
                                    <ResponsiveContainer width={"100%"} height={200}>
                                        <LineChart
                                            width={600}
                                            height={300}
                                            data={state.performanceEntries}
                                            syncId="anyId"
                                            margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                        >
                                            <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                            <XAxis dataKey="Checkpoint"/>
                                            <YAxis/>
                                            <Tooltip/>
                                            <Line type="monotone" dataKey="FPS" stroke="#ff7300" yAxisId={0}>
                                                <LabelList content={<CustomLineLabel unit={'fps'}/>}/>
                                            </Line>
                                        </LineChart>
                                    </ResponsiveContainer>
                                </Grid>
                                <Grid size={{xs: 12, md: 12}}>
                                    <Typography gutterBottom={true} variant="subtitle1">
                                        Memory
                                    </Typography>
                                    <Divider/>
                                    <ResponsiveContainer width={"100%"} height={200}>
                                        <LineChart
                                            width={600}
                                            height={300}
                                            data={state.performanceEntries}
                                            syncId="anyId"
                                            margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                        >
                                            <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                            <XAxis dataKey="Checkpoint"/>
                                            <YAxis/>
                                            <Tooltip/>
                                            <Line type="monotone" dataKey="MEM" stroke="#ff7300" yAxisId={0}>
                                                <LabelList content={<CustomLineLabel unit={'MB'}/>}/>
                                            </Line>
                                        </LineChart>
                                    </ResponsiveContainer>
                                </Grid>
                                <Grid size={{xs: 12, md: 12}}>
                                    <Typography gutterBottom={true} variant="subtitle1">
                                        CPU
                                    </Typography>
                                    <Divider/>
                                    <ResponsiveContainer width={"100%"} height={200}>
                                        <LineChart
                                            width={600}
                                            height={300}
                                            data={state.performanceEntries}
                                            syncId="anyId"
                                            margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                        >
                                            <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                            <XAxis dataKey="Checkpoint"/>
                                            <YAxis/>
                                            <Tooltip/>
                                            <Line type="monotone" dataKey="CPU" stroke="#ff7300" yAxisId={0}>
                                                <LabelList content={<CustomLineLabel unit={'%'}/>}/>
                                            </Line>
                                        </LineChart>
                                    </ResponsiveContainer>
                                </Grid>
                            </Grid>
                        </TabPanel>
                        {checkpoints !== undefined && checkpoints.length > 0 &&
                            <TabPanel value={value} index={5}>
                                <Grid container={true} sx={{padding: 1}} spacing={1}>
                                    <Grid size={{xs: 12, md: 12}}>
                                        <Typography gutterBottom={true} variant="subtitle1">
                                            Checkpoints
                                        </Typography>
                                        <Divider/>
                                        <ResponsiveContainer width={"100%"} height={200}>
                                            <LineChart
                                                width={600}
                                                height={300}
                                                margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                            >
                                                <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                                <XAxis xAxisId={'pruntime'} dataKey="Checkpoint"/>
                                                <XAxis xAxisId={'pexecutiontime'} dataKey="Checkpoint"
                                                       hide={true}/>
                                                <YAxis/>
                                                <Tooltip/>
                                                <Legend/>
                                                {
                                                    historicalCheckpoints.map((e, i) => (
                                                        <XAxis key={`hist_checkpoint_xaxis_${i}`}
                                                               xAxisId={`xachis_${i}`} dataKey="Checkpoint"
                                                               hide={true}/>))
                                                }
                                                {
                                                    historicalCheckpoints.map((e, i) => <Line
                                                        key={`hist_checkpoint_line_${i}`}
                                                        data={e.Performance} type="monotone" dataKey="Runtime"
                                                        stroke="#004467" strokeDasharray="5 5"
                                                        xAxisId={`xachis_${i}`}>
                                                        <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                                    </Line>)
                                                }
                                                <Line data={checkpoints} type="monotone" dataKey="Runtime"
                                                      stroke="#ff7300" xAxisId={'pruntime'}>
                                                    <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                                </Line>
                                                <Line data={checkpoints} type="monotone"
                                                      dataKey="ExecutionTime" stroke="#8884d8"
                                                      xAxisId={'pexecutiontime'}>
                                                    <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                                </Line>
                                            </LineChart>
                                        </ResponsiveContainer>
                                    </Grid>
                                </Grid>
                            </TabPanel>
                        }
                    </Grid>
                </Paper>
            </TitleCard>
        </Box>
    );
};

export default TestProtocolPage;
