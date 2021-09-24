import React, { ChangeEvent, FC, ReactElement, useEffect, useState } from 'react';
import { createStyles, Theme, WithStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';
import {
    Box,
    Button,
    FormControl,
    InputLabel,
    LinearProgress,
    LinearProgressProps,
    MenuItem,
    Typography,
    withStyles,
} from '@material-ui/core';
import IAppData from '../types/app';
import { getAllApps, uploadNewApp } from '../services/app.service';
import Select from '@material-ui/core/Select';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        root: {
            margin: 'auto',
        },
        paper: {
            width: 200,
            height: 230,
            overflow: 'auto',
        },
        button: {
            margin: theme.spacing(0.5, 0),
        },
        input: {
            display: 'none',
        },
        formControl: {
            margin: theme.spacing(1),
            minWidth: 120,
        },
    });


function LinearProgressWithLabel(props: LinearProgressProps & { value: number }): ReactElement {
    return (
        <Box sx={ { display: 'flex', alignItems: 'center' } }>
            <Box sx={ { width: '100%', mr: 1 } }>
                <LinearProgress variant="determinate" { ...props } />
            </Box>
            <Box sx={ { minWidth: 35 } }>
                <Typography variant="body2">{ `${ Math.round(
                    props.value,
                ) }%` }</Typography>
            </Box>
        </Box>
    );
}


interface AppSelectionProps extends WithStyles<typeof styles> {
    onSelectionChanged: (app: IAppData) => void;
    upload: boolean;
}

const AppSelection: FC<AppSelectionProps> = (props) => {
    const { classes, onSelectionChanged, upload } = props;

    const [app, setApp] = useState<IAppData>();
    const [apps, setApps] = useState<IAppData[]>([]);
    const [selectedAppID, setSelectedAppID] = useState<number>(1);

    const [uploadProgress, setUploadProgress] = useState<number>(0);

    useEffect(() => {
        getAllApps().then(response => {
            setApps(response.data);
        }).catch(e => {
        });
    }, []);

    useEffect(() => {
        if (app !== undefined) {
            onSelectionChanged(app);
        }
    }, [app, onSelectionChanged]);

    const selectUploadFile = (e: ChangeEvent<HTMLInputElement>): void => {
        e.preventDefault();
        if (e.target.files != null) {
            const f = e.target.files[ 0 ] as File;
            uploadNewApp(f, progress => {
                setUploadProgress(progress);
            }, data => {
                console.log(data.data);
                if (data.data !== undefined && data.data !== null) {
                    setApp(data.data);
                    setApps(prevState => {
                        const newState = [...prevState];
                        newState.push(data.data);
                        return newState;
                    });
                }
            });
        }
    };

    const handleChange = (e: ChangeEvent<{ name?: string | undefined, value: unknown }>): void => {
        e.preventDefault();
        console.log(e.target.value);
        console.log(e.target.name);
        if (e.target.value !== undefined) {
            const appId = e.target.value as number;
            const a = apps.find(element => element.ID == appId);
            setApp(a);
            setSelectedAppID(appId);
        }
    };

    return (
        <Grid
            container={ true }
            spacing={ 2 }
            justifyContent="center"
            alignItems="center"
            direction={ 'column' }
            className={ classes.root }
        >
            <Grid item={ true }>
                <FormControl className={ classes.formControl }>
                    <InputLabel id="demo-simple-select-label">App</InputLabel>
                    <Select
                        labelId="demo-simple-select-label"
                        id="demo-simple-select"
                        value={ selectedAppID }
                        onChange={ event => handleChange(event) }
                    >
                        { apps.map((a) =>
                            <MenuItem value={ a.ID }>{ a.Platform } { a.Name } ({ a.Version })</MenuItem>,
                        ) }
                    </Select>
                </FormControl>
            </Grid>
            <Grid item={ true }>
                <Grid
                    container={ true }
                    spacing={ 2 }
                    justifyContent="center"
                    alignItems="center"
                    className={ classes.root }
                >
                    { upload && uploadProgress > 0 && (
                    <Grid item={ true }>
                        <Box sx={ { width: '200px' } }>
                            <LinearProgressWithLabel value={ uploadProgress }/>
                        </Box>
                    </Grid>
                    ) }
                    { upload && (
                    <Grid item={ true }>
                        <input
                            accept="*.apk,*.ipa"
                            className={ classes.input }
                            id="app-upload"
                            multiple={ true }
                            type="file"
                            onChange={ (e) => {
                                selectUploadFile(e);
                            }
                            }/>
                        <label htmlFor="app-upload">
                            <Button variant="contained"
                                color="primary"
                                component="span">
                                Upload New
                            </Button>
                        </label>
                    </Grid>
                    ) }
                </Grid>
            </Grid>
        </Grid>
    );
};

export default withStyles(styles)(AppSelection);
