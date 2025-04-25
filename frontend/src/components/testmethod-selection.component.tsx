import React, {ChangeEvent, ReactElement, useEffect} from 'react';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Checkbox from '@mui/material/Checkbox';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import IAppFunctionData from '../types/app.function';
import {Alert, CircularProgress, Divider, TextField, Typography} from '@mui/material';
import {getTestFunctions} from '../services/unity.service';
import Grid from "@mui/material/Grid";
import {useError} from "../ErrorProvider";
import {Box} from "@mui/system";

interface TestMethodSelectionProps {
    selectedTestFunctions?: IAppFunctionData[]
    onSelectionChanged: (functions: IAppFunctionData[]) => void;
}

function not(a: IAppFunctionData[], b: IAppFunctionData[]): IAppFunctionData[] {
    return a.filter((value) => b.indexOf(value) === -1);
}

function intersection(a: IAppFunctionData[], b: IAppFunctionData[]): IAppFunctionData[] {
    return a.filter((value) => b.indexOf(value) !== -1);
}

const TestMethodSelection: React.FC<TestMethodSelectionProps> = (props) => {
    const {onSelectionChanged, selectedTestFunctions} = props;
    const {setError} = useError()

    const [state, setState] = React.useState<{
        loading: boolean,
        checked: IAppFunctionData[],
        testsJson: string,
        filterText: string
        left: IAppFunctionData[],
        right: IAppFunctionData[],
    }>({
        loading: true,
        checked: [],
        testsJson: '',
        filterText: '',
        left: [],
        right: selectedTestFunctions ? selectedTestFunctions : [],
    })

    const leftChecked = intersection(state.checked, state.left);
    const rightChecked = intersection(state.checked, state.right);

    const changeJsonFunction = (event: ChangeEvent<HTMLTextAreaElement>): void => {
        if (event.target.value !== null && event.target.value !== '') {
            setState(prevState => ({...prevState, testsJson: event.target.value}));
        }
    };

    function trimClass(className: string): string {
        const elements = className.split('.');
        return elements[elements.length - 1];
    }

    function trimMethod(methodSignature: string): string {
        const elements = methodSignature.split(' ');
        return elements[elements.length - 1];
    }

    const applyFilter = (apps: IAppFunctionData[], filter: string): IAppFunctionData[] => {
        filter = filter.toLowerCase()
        return apps.filter(a => a.method.toLowerCase().indexOf(filter) >= 0 || a.class.toLowerCase().indexOf(filter) >= 0 || a.assembly.toLowerCase().indexOf(filter) >= 0)
    }
    const leftFiltered = applyFilter(state.left, state.filterText);

    const handleToggle = (value: IAppFunctionData) => () => {
        const currentIndex = state.checked.indexOf(value);
        const newChecked = [...state.checked];

        if (currentIndex === -1) {
            newChecked.push(value);
        } else {
            newChecked.splice(currentIndex, 1);
        }

        setState(prevState => ({...prevState, checked: newChecked}));
    };

    const handleAllRight = (): void => {
        if (state.filterText.length > 0) {
            const filtered = applyFilter(state.left, state.filterText);
            setState(prevState => ({
                ...prevState,
                left: not(prevState.left, filtered),
                right: prevState.right.concat(filtered)
            }));
        } else {
            setState(prevState => ({...prevState, left: [], right: prevState.right.concat(prevState.left)}));
        }
    };

    const handleCheckedRight = (): void => {
        setState(prevState => ({
            ...prevState,
            left: not(prevState.left, leftChecked),
            right: prevState.right.concat(leftChecked),
            checked: not(prevState.checked, leftChecked)
        }));
    };

    const handleCheckedLeft = (): void => {
        setState(prevState => ({
            ...prevState,
            left: prevState.left.concat(rightChecked),
            right: not(prevState.right, rightChecked),
            checked: not(prevState.checked, rightChecked)
        }));
    };

    const handleAllLeft = (): void => {
        setState(prevState => ({...prevState, left: prevState.left.concat(prevState.right), right: []}));
    };

    useEffect(() => {
        if (state.testsJson !== '') {
            const testFunctions: IAppFunctionData[] = JSON.parse(state.testsJson);
            setState(prevState => ({...prevState, left: not(testFunctions, prevState.right)}));
        }
    }, [state.testsJson]);

    useEffect(() => {
        onSelectionChanged(state.right);
    }, [state.right, onSelectionChanged]);


    const customList = (items: IAppFunctionData[], showfilter: boolean): ReactElement => (
        <>
            <Box sx={{height: 56}}>
                {showfilter && <TextField sx={{padding: 1}} placeholder={'Filter Functions'} value={state.filterText}
                                          fullWidth={true} size={"small"}
                                          onChange={onChangeFilter}/>}
            </Box>
            <Divider/>
            <Paper sx={{minWidth: 400, minHeight: 500, maxHeight: 500, margin: 'auto', overflow: 'auto'}}>
                <List dense={true} component="div" role="list">
                    {items.map((value) => {
                        const labelId = `transfer-list-item-${value.class}-${value.method}-label`;

                        return (
                            <ListItem key={value.class + value.method} role="listitem" onClick={handleToggle(value)}>
                                <ListItemIcon>
                                    <Checkbox
                                        checked={state.checked.indexOf(value) !== -1}
                                        tabIndex={-1}
                                        disableRipple={true}
                                        inputProps={{'aria-labelledby': labelId}}
                                    />
                                </ListItemIcon>
                                <ListItemText id={labelId}
                                              primary={`${trimMethod(value.method)}`}
                                              secondary={trimClass(value.class)}/>
                            </ListItem>
                        );
                    })}
                    <ListItem/>
                </List>
            </Paper>
        </>
    );

    useEffect(() => {
        setState(prevState => ({...prevState, loading: true}));
        getTestFunctions().then(tfs => {
            setState(prevState => ({
                ...prevState,
                left: tfs.filter(d => prevState.right.findIndex(d1 => d.assembly === d.assembly && d.class === d1.class && d.method === d1.method) < 0)
            }));
        }).catch(ex => setError(ex)).finally(() => {
            setState(prevState => ({...prevState, loading: false}));
        });
    }, []);

    const onChangeFilter = (event: ChangeEvent<HTMLTextAreaElement>) => {
        setState(prevState => ({...prevState, filterText: event.target.value}));
    }

    return (
        state.loading ? (<><Typography variant={'h6'}>Try to connect to your local Unity Editor to fetch
            Tests...</Typography><CircularProgress
            color="inherit"/></>) : (
            <Grid container={true}
                  spacing={2}
                  justifyContent={'center'}
                  alignItems={'center'}
                  direction={'column'}>
                {(state.left.length == 0 && state.right.length == 0 && (
                    <Grid container={true}>
                        <Grid size={12}>
                            <Alert severity="info">Unable to Connect to your local Unity Editor or no tests found please
                                check your settings in Unity "Tools" - "Automation Hub" - "Settings"<br/>You can also
                                copy
                                the tests json from unity and paste them here.</Alert>
                        </Grid>
                        <Grid size={12}>
                            <TextField
                                id="outlined-multiline-static"
                                placeholder="Paste Test Functions here"
                                multiline={true}
                                fullWidth={true}
                                rows={6}
                                defaultValue={state.testsJson}
                                variant="outlined"
                                onChange={changeJsonFunction}
                            />
                        </Grid>
                    </Grid>))
                }
                <Grid container={true} size={12}>
                    <Grid
                        container={true}
                        spacing={2}
                        justifyContent="center"
                        alignItems="center"
                    >
                        <Grid spacing={2} container={true} size={5}>
                            <Grid size={12}>
                                <Typography variant={'subtitle1'}>Available Test Functions</Typography>
                            </Grid>
                            <Grid size={12}>
                                {/*<TextField id={'filter'} onChange={onChangeFilter} value={filterText}
                                            fullWidth={true}/> */}
                            </Grid>
                            <Grid size={12}>
                                {customList(applyFilter(leftFiltered, state.filterText), true)}
                            </Grid>
                        </Grid>
                        <Grid spacing={2} container={true} size={1}>
                            <Grid container={true} direction="column" alignItems="center">
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={handleAllRight}
                                    disabled={leftFiltered.length === 0}
                                    aria-label="move all right"
                                >
                                    ≫
                                </Button>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={handleCheckedRight}
                                    disabled={leftChecked.length === 0}
                                    aria-label="move selected right"
                                >
                                    &gt;
                                </Button>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={handleCheckedLeft}
                                    disabled={rightChecked.length === 0}
                                    aria-label="move selected left"
                                >
                                    &lt;
                                </Button>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={handleAllLeft}
                                    disabled={state.right.length === 0}
                                    aria-label="move all left"
                                >
                                    ≪
                                </Button>
                            </Grid>
                        </Grid>
                        <Grid spacing={2} container={true} size={5}>
                            <Grid size={12}>
                                <Typography variant={'subtitle1'}>Selected Test Functions</Typography><br/>
                            </Grid>
                            <Grid size={12}>
                            </Grid>
                            <Grid size={12}>
                                {customList(state.right, false)}
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
            </Grid>)
    );
};

export default TestMethodSelection;

function onChangeFilter(event: ChangeEvent<HTMLTextAreaElement | HTMLInputElement>): void {
    throw new Error('Function not implemented.');
}

