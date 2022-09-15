import React, { ChangeEvent, ReactElement, useEffect } from 'react';
import Grid from '@mui/material/Grid';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Checkbox from '@mui/material/Checkbox';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import IAppFunctionData from '../types/app.function';
import { Alert, CircularProgress, TextField, Typography } from '@mui/material';
import { getTestFunctions } from '../services/unity.service';

interface TestMethodSelectionProps {
    onSelectionChanged: (functions: IAppFunctionData[]) => void;
}

function not(a: IAppFunctionData[], b: IAppFunctionData[]): IAppFunctionData[] {
    return a.filter((value) => b.indexOf(value) === -1);
}

function intersection(a: IAppFunctionData[], b: IAppFunctionData[]): IAppFunctionData[] {
    return a.filter((value) => b.indexOf(value) !== -1);
}

const TestMethodSelection: React.FC<TestMethodSelectionProps> = (props) => {
    const { onSelectionChanged } = props;

    const [loading, setLoading] = React.useState<boolean>(true);

    const [checked, setChecked] = React.useState<IAppFunctionData[]>([]);

    const [testJson, setTestJson] = React.useState<string>('');

    const [left, setLeft] = React.useState<IAppFunctionData[]>([]);
    const [right, setRight] = React.useState<IAppFunctionData[]>([]);

    const leftChecked = intersection(checked, left);
    const rightChecked = intersection(checked, right);

    const handleToggle = (value: IAppFunctionData) => () => {
        const currentIndex = checked.indexOf(value);
        const newChecked = [...checked];

        if (currentIndex === -1) {
            newChecked.push(value);
        } else {
            newChecked.splice(currentIndex, 1);
        }

        setChecked(newChecked);
    };

    const handleAllRight = (): void => {
        setRight(right.concat(left));
        setLeft([]);
    };

    const handleCheckedRight = (): void => {
        setRight(right.concat(leftChecked));
        setLeft(not(left, leftChecked));
        setChecked(not(checked, leftChecked));
    };

    const handleCheckedLeft = (): void => {
        setLeft(left.concat(rightChecked));
        setRight(not(right, rightChecked));
        setChecked(not(checked, rightChecked));
    };

    const handleAllLeft = (): void => {
        setLeft(left.concat(right));
        setRight([]);
    };

    useEffect(() => {
        if (testJson !== '') {
            const testFunctions: IAppFunctionData[] = JSON.parse(testJson);
            setLeft(testFunctions);
            setRight([]);
        }
    }, [testJson]);

    useEffect(() => {
        onSelectionChanged(right);
    }, [right, onSelectionChanged]);

    const changeJsonFunction = (event: ChangeEvent<HTMLTextAreaElement>): void => {
        if (event.target.value !== null && event.target.value !== '') {
            setTestJson(event.target.value);
        }
    };

    function trimClass(className: string): string {
        const elements = className.split('.');
        return elements[ elements.length - 1 ];
    }

    function trimMethod(methodSignature: string): string {
        const elements = methodSignature.split(' ');
        return elements[ elements.length - 1 ];
    }

    const customList = (items: IAppFunctionData[]): ReactElement => (
        <Paper sx={ { minWidth: 400, maxWidth: 400, minHeight: 400, maxHeight: 400, margin: 'auto', overflow: 'auto' } }>
            <List dense={ true } component="div" role="list">
                { items.map((value) => {
                    const labelId = `transfer-list-item-${ value.Class }-${ value.Method }-label`;

                    return (
                        <ListItem key={ value.Class + value.Method } role="listitem" button={ true } onClick={ handleToggle(value) }>
                            <ListItemIcon>
                                <Checkbox
                                    checked={ checked.indexOf(value) !== -1 }
                                    tabIndex={ -1 }
                                    disableRipple={ true }
                                    inputProps={ { 'aria-labelledby': labelId } }
                                />
                            </ListItemIcon>
                            <ListItemText id={ labelId }
                                primary={ `${ trimMethod(value.Method) }` } secondary={trimClass(value.Class) }/>
                        </ListItem>
                    );
                }) }
                <ListItem/>
            </List>
        </Paper>
    );

    useEffect(() => {
        setLoading(true);
        getTestFunctions().then(response => {
            setLeft(response.data);
            setRight([]);
        }).finally(() => {
            setLoading(false);
        });
    }, []);

    return (
        loading ? (<><Typography variant={ 'h6' }>Try to connect to your local Unity Editor to fetch Tests...</Typography><CircularProgress color="inherit" /></>) : (
            <Grid container={ true }
                spacing={ 2 }
                justifyContent={ 'center' }
                alignItems={ 'center' }
                direction={ 'column' }>
                { (left.length == 0 && right.length == 0 && (
                    <Grid item={ true }>
                        <Alert severity="info">Unable to Connect to your local Unity Editor or no tests found please check your settings in Unity "Tools" - "Automation Hub" - "Settings"<br />You can also copy the tests json from unity and paste them here.</Alert><br/>
                        <TextField
                            id="outlined-multiline-static"
                            label="Paste Test Functions here"
                            multiline={ true }
                            rows={ 6 }
                            defaultValue={ testJson }
                            variant="outlined"
                            onChange={ changeJsonFunction }
                        />
                    </Grid>))
                }
                <Grid item={ true }>
                    <Grid
                        container={ true }
                        spacing={ 2 }
                        justifyContent="center"
                        alignItems="center"
                    >

                        <Grid item={ true }>
                            <Typography variant={ 'subtitle1' }>Available Test Functions</Typography><br/>
                            { customList(left) }
                        </Grid>
                        <Grid item={ true }>
                            <Grid container={ true } direction="column" alignItems="center">
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={ handleAllRight }
                                    disabled={ left.length === 0 }
                                    aria-label="move all right"
                                >
                                    ≫
                                </Button>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={ handleCheckedRight }
                                    disabled={ leftChecked.length === 0 }
                                    aria-label="move selected right"
                                >
                                    &gt;
                                </Button>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={ handleCheckedLeft }
                                    disabled={ rightChecked.length === 0 }
                                    aria-label="move selected left"
                                >
                                    &lt;
                                </Button>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    onClick={ handleAllLeft }
                                    disabled={ right.length === 0 }
                                    aria-label="move all left"
                                >
                                    ≪
                                </Button>
                            </Grid>
                        </Grid>
                        <Grid item={ true }>
                            <Typography variant={ 'subtitle1' }>Selected Test Functions</Typography><br/>
                            { customList(right) }
                        </Grid>
                    </Grid>
                </Grid>
            </Grid>)
    );
};

export default TestMethodSelection;
