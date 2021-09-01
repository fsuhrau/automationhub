import React, {useState} from 'react';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import {createStyles, Theme, withStyles, WithStyles} from '@material-ui/core/styles';
import Select, { SelectChangeEvent } from '@material-ui/core/Select';
import {Box, FormControl, InputLabel, MenuItem, Step, StepLabel, Stepper, TextField} from "@material-ui/core";
import ITestData from "../../types/test";
import Divider from "@material-ui/core/Divider";

const styles = (theme: Theme) =>
    createStyles({
        paper: {
            maxWidth: 936,
            margin: 'auto',
            overflow: 'hidden',
        },
        searchBar: {
            borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
        },
        searchInput: {
            fontSize: theme.typography.fontSize,
        },
        block: {
            display: 'block',
        },
        addUser: {
            marginRight: theme.spacing(1),
        },
        contentWrapper: {
            margin: '40px 16px',
        },
    });

const steps = ['Test', 'Upload App', 'Scenario'];
const types = [{id: 1, name: 'Selecnium'}, {id: 2, name: 'Cocos'}, {id: 3, name: 'Unity'}];

export interface AddTestProps extends WithStyles<typeof styles> {
}

function AddTestPage(props: AddTestProps) {
    const {classes} = props;

    const [test, setTest] = useState<ITestData>();
    const [type, setType] = useState('');

    const [age, setAge] = React.useState('');

    const handleChange = (event: SelectChangeEvent) => {
        setType(event.target.value as string);
    };

    return (
        <Paper className={classes.paper}>
            <AppBar className={classes.searchBar} position="static" color="default" elevation={0}>
                <Toolbar>
                    <Grid container spacing={2} alignItems="center">
                        Create New Test
                    </Grid>
                </Toolbar>
            </AppBar>
            <Stepper activeStep={1} alternativeLabel>
                {steps.map((label) => (
                    <Step key={label}>
                        <StepLabel>{label}</StepLabel>
                    </Step>
                ))}
            </Stepper>
            <Divider variant="middle"/>
            <Box
                component="form"
                sx={{
                    '& .MuiTextField-root': {m: 1, width: '25ch'},
                }}
                noValidate
                autoComplete="off"
            >
                <FormControl>
                    <InputLabel id="test-name-input-label">Name</InputLabel>
                    <TextField id="test-name-input" label="test-name-input-label" variant="standard"/>
                </FormControl>
                <br />
                <FormControl>
                    <InputLabel id="select-test-type-label">Type</InputLabel>
                    <Select
                        labelId="select-test-type-label"
                        id="select-test-type"
                        value={type}
                        label="Type"
                        onChange={handleChange}
                    >
                        {types.map((label) => (
                            <MenuItem value={label.id}>{label.name}</MenuItem>
                        ))}
                    </Select>
                </FormControl>
            </Box>
        </Paper>
    );
}

export default withStyles(styles)(AddTestPage);