import React, { ChangeEvent, FC, ReactElement, useEffect } from 'react';
import Grid from '@mui/material/Grid';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Checkbox from '@mui/material/Checkbox';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import IAppFunctionData from '../types/app.function';
import { TextField, Typography } from '@mui/material';
import { getTestFunctions } from '../services/unity.service';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles(theme => ({
    root: {
        margin: 'auto',
    },
    paper: {
        width: 300,
        height: 330,
        overflow: 'auto',
    },
    button: {
        margin: '1px',
    },
}));

interface TestMethodSelectionProps {
    onSelectionChanged: (functions: IAppFunctionData[]) => void;
}

function not(a: IAppFunctionData[], b: IAppFunctionData[]): IAppFunctionData[] {
    return a.filter((value) => b.indexOf(value) === -1);
}

function intersection(a: IAppFunctionData[], b: IAppFunctionData[]): IAppFunctionData[] {
    return a.filter((value) => b.indexOf(value) !== -1);
}

const TestMethodSelection: FC<TestMethodSelectionProps> = (props) => {
    const classes = useStyles();
    const { onSelectionChanged } = props;

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
        <Paper className={ classes.paper } style={ { overflow: 'auto' } }>
            <List dense={ true } component="div" role="list">
                { items.map((value) => {
                    const labelId = `transfer-list-item-${ value.ID }-label`;

                    return (
                        <ListItem key={ value.ID } role="listitem" button={ true } onClick={ handleToggle(value) }>
                            <ListItemIcon>
                                <Checkbox
                                    checked={ checked.indexOf(value) !== -1 }
                                    tabIndex={ -1 }
                                    disableRipple={ true }
                                    inputProps={ { 'aria-labelledby': labelId } }
                                />
                            </ListItemIcon>
                            <ListItemText id={ labelId }
                                primary={ `${ trimClass(value.Class) } ${ trimMethod(value.Method) }` }/>
                        </ListItem>
                    );
                }) }
                <ListItem/>
            </List>
        </Paper>
    );

    useEffect(() => {
        getTestFunctions().then(response => {
            setLeft(response.data);
            setRight([]);
        });
    }, []);

    return (
        <Grid container={ true }
            spacing={ 2 }
            justifyContent={ 'center' }
            alignItems={ 'center' }
            direction={ 'column' }>
            <Grid item={ true }>
                <Typography variant={ 'h6' }>Copy Test Export from Unity Editor.</Typography><br/>
                <TextField
                    id="outlined-multiline-static"
                    label="Paste Test Functions here"
                    multiline={ true }
                    rows={ 6 }
                    defaultValue={ testJson }
                    variant="outlined"
                    onChange={ changeJsonFunction }
                />
            </Grid>
            <Grid item={ true }>
                <Grid
                    container={ true }
                    spacing={ 2 }
                    justifyContent="center"
                    alignItems="center"
                    className={ classes.root }
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
                                className={ classes.button }
                                onClick={ handleAllRight }
                                disabled={ left.length === 0 }
                                aria-label="move all right"
                            >
                                ≫
                            </Button>
                            <Button
                                variant="outlined"
                                size="small"
                                className={ classes.button }
                                onClick={ handleCheckedRight }
                                disabled={ leftChecked.length === 0 }
                                aria-label="move selected right"
                            >
                                &gt;
                            </Button>
                            <Button
                                variant="outlined"
                                size="small"
                                className={ classes.button }
                                onClick={ handleCheckedLeft }
                                disabled={ rightChecked.length === 0 }
                                aria-label="move selected left"
                            >
                                &lt;
                            </Button>
                            <Button
                                variant="outlined"
                                size="small"
                                className={ classes.button }
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
        </Grid>
    );
};

export default TestMethodSelection;
