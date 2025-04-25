import React, {FunctionComponent, ReactElement, useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import ITestRunData from '../../types/test.run';
import {TestResultState} from '../../types/test.result.state.enum';
import ITestProtocolData, {duration} from '../../types/test.protocol';
import Typography from '@mui/material/Typography';
import {AvTimer, DateRange, PhoneAndroid, Speed} from '@mui/icons-material';
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
import Grid from "@mui/material/Grid";
import {useLocation} from "react-router-dom";
import {byteFormat, fixedTwoFormat, kFormat} from "./value_formatter";

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
    testProtocolId: number,
    entry: IProtocolEntryData,
}

interface TestProtocolPageProps {
    run: ITestRunData
    protocol: ITestProtocolData
}

const CustomLineLabel: FunctionComponent<any> = (props: any) => {
    const {x, y, stroke, value, unit} = props;
    return (
        <text x={x} y={y} dy={-4} fill={stroke} fontSize={10} textAnchor="middle">
            {value?.toFixed(2)}{unit !== undefined ? unit : ''}
        </text>
    );
};

interface PerformanceNoteLabelProps {
    value: number,
    isHigherBetter: boolean,
    formatter?: (value: number) => string
}

const PerformanceNoteLabel: React.FC<PerformanceNoteLabelProps> = (props: PerformanceNoteLabelProps) => {
    const {value, isHigherBetter, formatter} = props;
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
    let formattedValue: string;
    if (formatter) {
        formattedValue = formatter(value)
    } else {
        formattedValue = value.toString()
    }
    return (
        <span style={{
            color: color,
            fontSize: '0.8em',
            verticalAlign: "top"
        }}>{value >= 0 ? '+' : ''}{formattedValue}</span>
    );
};

const tabs = ['status', 'log', 'screenshots', 'video', 'performance', 'checkpoint'];

const useQuery = () => {
    return new URLSearchParams(useLocation().search);
};

const useUpdateQueryParam = (key: string, value: string) => {
    const url = new URL(window.location.href);
    url.searchParams.set(key, value);
    window.history.pushState({}, '', url.toString());
};

const TestProtocolPage: React.FC<TestProtocolPageProps> = (props) => {

    const {protocol} = props;

    const query = useQuery();

    const tabIndexOrDefault = (t: string | null): number => {
        return t && tabs.includes(t) ? tabs.indexOf(t) : 0;
    }

    const [value, setValue] = useState(tabIndexOrDefault(query.get('t')));

    const handleChange = (event: React.ChangeEvent<{}>, newValue: number): void => {
        setValue(newValue);
        useUpdateQueryParam('t', tabs[newValue])
    };

    // handle browser back button with query parameter
    useEffect(() => {
        const handlePopState = () => {
            const newQuery = new URLSearchParams(window.location.search);
            setValue(tabIndexOrDefault(newQuery.get('t')))
        };

        window.addEventListener('popstate', handlePopState);

        return () => {
            window.removeEventListener('popstate', handlePopState);
        };
    }, []);

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
        steps: number,
        checkpoints: IProtocolPerformanceEntryData[],
        history: ITestProtocolData[],
    }>({
        lastScreen: null,
        lastErrors: [],
        lastStep: null,
        entries: [],
        screenEntries: [],
        performanceEntries: [],
        steps: 0,
        checkpoints: [],
        history: [],
    })

    const protocolEntry = useSSE<NewTestProtocolLogPayload | null>(`test_protocol_${protocol.id}_log`, null);
    useEffect(() => {
        if (protocolEntry === null)
            return;
        setState(prevState => ({...prevState, entries: [...prevState.entries, protocolEntry.entry]}))
    }, [protocolEntry]);

    const updateStatusEntries = (): void => {
        const length = state.entries.length;
        let numSteps = 0;
        const errors: IProtocolEntryData[] = [];
        let lastStep: IProtocolEntryData | null = null;
        let lastScreen: IProtocolEntryData | null = null;
        const screenEntries: IProtocolEntryData[] = [];
        for (let i = length - 1; i > 0; i--) {
            if (state.entries[i].source === 'screen') {
                screenEntries.push(state.entries[i])
                if (lastScreen === null) {
                    lastScreen = state.entries[i];
                }
            }
            if (state.entries[i].source === 'step') {
                numSteps++;
                if (lastStep === null) {
                    lastStep = state.entries[i];
                }
            }
            if (state.entries[i].level === 'error') {
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

    const getCheckpoints = (entries: IProtocolPerformanceEntryData[]) => {
        if (entries == null) {
            return []
        }
        return entries.filter(p => p.checkpoint !== "schedule").map((value1, index, array) => {
            if (index !== 0) {
                value1.executionTime = value1.runtime - array[index - 1].runtime
            } else {
                value1.executionTime = value1.runtime
            }
            return value1
        })
    }

    const getHistoricalCheckpoints = (testProtocols: ITestProtocolData[] | null) => {
        if (testProtocols == null) {
            return []
        }
        return testProtocols.slice(-2).map(p => {
            p.performance = p.performance?.filter(p => p.checkpoint !== "schedule").map((value1, index, array) => {
                if (index !== 0) {
                    value1.executionTime = value1.runtime - array[index - 1].runtime
                } else {
                    value1.executionTime = value1.runtime
                }
                return value1
            })
            return p
        })
    }

    useEffect(() => {
        setState(prevState => ({
            ...prevState,
            entries: protocol.entries,
            performanceEntries: protocol.performance ? protocol.performance : [],
            checkpoints: getCheckpoints(protocol.performance),
            history: getHistoricalCheckpoints(protocol.testProtocolHistory)
        }))
    }, [protocol]);

    useEffect(() => {
        updateStatusEntries();
    }, [state.entries]);

    const diffAvgFPS = protocol.avgFps - protocol.histAvgFps;
    const diffAvgMEM = protocol.avgMem - protocol.histAvgMem;
    const diffAvgCPU = protocol.avgCpu - protocol.histAvgCpu;
    const diffAvgVertexCount = protocol.avgVertexCount - protocol.histAvgVertexCount;
    const diffAvgTriangles = protocol.avgTriangles - protocol.histAvgTriangles;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={protocol.testName}>
                <Grid container={true} spacing={1} alignItems="center" sx={{
                    padding: 1,
                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                }}>
                    <Grid>
                        <TestStatusIconComponent
                            status={protocol.testResult}/>
                    </Grid>
                    <Grid>
                        <DateRange sx={{display: 'block'}} color="inherit"/>
                    </Grid>
                    <Grid>
                        {moment(protocol.startedAt).format('YYYY/MM/DD HH:mm:ss')}
                    </Grid>

                    <Grid>
                        <AvTimer sx={{display: 'block'}} color="inherit"/>
                    </Grid>

                    <Grid>
                        {duration(protocol.startedAt, protocol.endedAt)}
                    </Grid>

                    {protocol.device && <Grid>
                        <PhoneAndroid sx={{display: 'block'}} color="inherit"/>
                    </Grid>}
                    {protocol.device && <Grid>
                        {protocol.device && (protocol.device.alias.length > 0 ? protocol.device.alias : protocol.device.name)}
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
                                {protocol?.testResult == TestResultState.TestResultFailed &&
                                    <Grid container={true} sx={{padding: 1, backgroundColor: '#ff2b40'}}
                                          size={{xs: 12, md: 12}}>
                                        <Grid container={true} size={{xs: 12, md: 12}} spacing={2}>
                                            {state.lastStep &&
                                                <Grid container={true} size={12}>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography
                                                            variant={"body2"}>{moment(state.lastStep.createdAt).format('YYYY/MM/DD HH:mm:ss')}</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}}>
                                                        <Typography
                                                            variant={"body2"}>{state.lastStep.message}</Typography>
                                                    </Grid>
                                                    <Grid size={12}>
                                                        <Divider/>
                                                    </Grid>
                                                </Grid>
                                            }
                                            {state.lastErrors.map((lastError, index) =>
                                                (<Grid key={`last_error_${index}`} container={true} size={12}>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography
                                                            variant={"body2"}>{moment(lastError.createdAt).format('YYYY/MM/DD HH:mm:ss')}</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}}>
                                                        <Typography
                                                            variant={"body2"}>{lastError.message}</Typography>
                                                    </Grid>
                                                    <Grid size={12}>
                                                        <Divider/>
                                                    </Grid>
                                                </Grid>)
                                            )}
                                        </Grid>
                                        <Grid size={{xs: 12, md: 12}} container={true}
                                              justifyContent={"center"} alignItems={"center"} justifyItems={"center"}>
                                            {state.lastScreen && <div>
                                                <Button aria-describedby={'last_screen_' + state.lastScreen.id}
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
                                                            image={`/api/data/${state.lastScreen.data}`}
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
                                                {duration(protocol.startedAt, protocol.endedAt)}
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
                                                    {protocol.avgFps?.toFixed(2)} <PerformanceNoteLabel
                                                    value={diffAvgFPS}
                                                    isHigherBetter={true}
                                                    formatter={fixedTwoFormat}
                                                />
                                                </Typography>
                                            </Grid>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                    Memory
                                                </Typography>
                                                <Typography variant="h5" component="h5">
                                                    {byteFormat(protocol.avgMem * 1024 * 1024)} <PerformanceNoteLabel
                                                    value={diffAvgMEM * 1024 * 1024}
                                                    isHigherBetter={false}
                                                    formatter={byteFormat}
                                                />
                                                </Typography>
                                            </Grid>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                    CPU
                                                </Typography>
                                                <Typography variant="h5" component="h5">
                                                    {protocol.avgCpu?.toFixed(2)}% <PerformanceNoteLabel
                                                    value={diffAvgCPU}
                                                    isHigherBetter={false}
                                                    formatter={fixedTwoFormat}
                                                />
                                                </Typography>
                                            </Grid>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                    VertexCount
                                                </Typography>
                                                <Typography variant="h5" component="h5">
                                                    {kFormat(protocol.avgVertexCount)} <PerformanceNoteLabel
                                                    value={diffAvgVertexCount}
                                                    isHigherBetter={false}
                                                    formatter={kFormat}
                                                />
                                                </Typography>
                                            </Grid>
                                            <Grid>
                                                <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                    Triangles
                                                </Typography>
                                                <Typography variant="h5" component="h5">
                                                    {kFormat(protocol.avgTriangles)} <PerformanceNoteLabel
                                                    value={diffAvgTriangles}
                                                    isHigherBetter={false}
                                                    formatter={kFormat}
                                                />
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
                                            data={state.performanceEntries === null ? [] : state.performanceEntries}
                                            syncId="anyId"
                                            margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                        >
                                            <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                            <XAxis dataKey="checkpoint"/>
                                            <YAxis/>
                                            <Tooltip contentStyle={{color: 'primary', backgroundColor: 'background'}} formatter={value1 => fixedTwoFormat(value1 as number)}/>
                                            <Line type="monotone" dataKey="fps" stroke="#ff7300" yAxisId={0}>
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
                                            data={state.performanceEntries == null ? [] : state.performanceEntries}
                                            syncId="anyId"
                                            margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                        >
                                            <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                            <XAxis dataKey="checkpoint"/>
                                            <YAxis/>
                                            <Tooltip contentStyle={{color: 'primary', backgroundColor: 'background'}} formatter={value1 => byteFormat((value1 as number) * 1024 * 1024)}/>
                                            <Line type="monotone" dataKey="mem" stroke="#ff7300" yAxisId={0}>
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
                                            data={state.performanceEntries == null ? [] : state.performanceEntries}
                                            syncId="anyId"
                                            margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                        >
                                            <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                            <XAxis dataKey="checkpoint"/>
                                            <YAxis/>
                                            <Tooltip contentStyle={{color: 'primary', backgroundColor: 'background'}}/>
                                            <Line type="monotone" dataKey="cpu" stroke="#ff7300" yAxisId={0}>
                                                <LabelList content={<CustomLineLabel unit={'%'}/>}/>
                                            </Line>
                                        </LineChart>
                                    </ResponsiveContainer>
                                </Grid>
                            </Grid>
                            <Grid size={{xs: 12, md: 12}}>
                                <Typography gutterBottom={true} variant="subtitle1">
                                    VertexCount
                                </Typography>
                                <Divider/>
                                <ResponsiveContainer width={"100%"} height={200}>
                                    <LineChart
                                        width={600}
                                        height={300}
                                        data={state.performanceEntries == null ? [] : state.performanceEntries}
                                        syncId="anyId"
                                        margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                    >
                                        <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                        <XAxis dataKey="checkpoint"/>
                                        <YAxis/>
                                        <Tooltip contentStyle={{color: 'primary', backgroundColor: 'background'}}/>
                                        <Line type="monotone" dataKey="vertexCount" stroke="#ff7300" yAxisId={0}>
                                            <LabelList content={<CustomLineLabel unit={''}/>}/>
                                        </Line>
                                    </LineChart>
                                </ResponsiveContainer>
                            </Grid>
                            <Grid size={{xs: 12, md: 12}}>
                                <Typography gutterBottom={true} variant="subtitle1">
                                    Triangles
                                </Typography>
                                <Divider/>
                                <ResponsiveContainer width={"100%"} height={200}>
                                    <LineChart
                                        width={600}
                                        height={300}
                                        data={state.performanceEntries == null ? [] : state.performanceEntries}
                                        syncId="anyId"
                                        margin={{top: 10, right: 20, left: 10, bottom: 5}}
                                    >
                                        <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                        <XAxis dataKey="checkpoint"/>
                                        <YAxis/>
                                        <Tooltip contentStyle={{color: 'primary', backgroundColor: 'background'}}/>
                                        <Line type="monotone" dataKey="triangles" stroke="#ff7300" yAxisId={0}>
                                            <LabelList content={<CustomLineLabel unit={''}/>}/>
                                        </Line>
                                    </LineChart>
                                </ResponsiveContainer>
                            </Grid>
                        </TabPanel>
                        {state.checkpoints.length > 0 &&
                            <TabPanel value={value} index={5}>
                                <Grid container={true} sx={{padding: 1}} spacing={1}>
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
                                            <XAxis xAxisId={'pruntime'} dataKey="checkpoint"/>
                                            <XAxis xAxisId={'pexecutiontime'} dataKey="checkpoint"
                                                   hide={true}/>
                                            <YAxis/>
                                            <Tooltip
                                                contentStyle={{color: 'primary', backgroundColor: 'background'}}/>
                                            <Legend/>
                                            {
                                                state.history.map((e, i) => (
                                                    <XAxis key={`hist_checkpoint_xaxis_${i}`}
                                                           xAxisId={`xachis_${i}`} dataKey="checkpoint"
                                                           hide={true}/>))
                                            }
                                            {
                                                state.history.map((e, i) => <Line
                                                    key={`hist_checkpoint_line_${i}`}
                                                    data={e.performance} type="monotone" dataKey="runtime"
                                                    stroke="#004467" strokeDasharray="5 5"
                                                    xAxisId={`xachis_${i}`}>
                                                    <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                                </Line>)
                                            }
                                            <Line data={state.checkpoints} type="monotone" dataKey="runtime"
                                                  stroke="#ff7300" xAxisId={'pruntime'}>
                                                <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                            </Line>
                                            <Line data={state.checkpoints} type="monotone"
                                                  dataKey="executionTime" stroke="#8884d8"
                                                  xAxisId={'pexecutiontime'}>
                                                <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                            </Line>
                                        </LineChart>
                                    </ResponsiveContainer>
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
