import { ChangeEvent, FC, ReactElement, useEffect, useState } from 'react';
import Grid from '@mui/material/Grid';
import {
    Box,
    Button,
    FormControl,
    InputLabel,
    LinearProgress,
    LinearProgressProps,
    MenuItem, SelectChangeEvent,
    Typography,
} from '@mui/material';
import IAppData from '../types/app';
import { getAllApps, uploadNewApp } from '../services/app.service';
import Select from '@mui/material/Select';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles(theme => ({
    root: {
        margin: 'auto',
    },
    paper: {
        width: 200,
        height: 230,
        overflow: 'auto',
    },
    input: {
        display: 'none',
    },
    formControl: {
        margin: '1px',
        minWidth: 120,
    },
}));


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


interface AppSelectionProps {
    onSelectionChanged: (app: IAppData) => void;
    upload: boolean;
}

const AppSelection: FC<AppSelectionProps> = (props) => {
    const classes = useStyles();
    const { onSelectionChanged, upload } = props;

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
                if (data.data !== undefined && data.data !== null) {
                    setApps(prevState => {
                        const newState = [...prevState];
                        newState.push(data.data);
                        return newState;
                    });
                    setApp(data.data);
                    setSelectedAppID(data.data.ID);
                }
            });
        }
    };

    const handleChange = (e: SelectChangeEvent<number>): void => {
        e.preventDefault();
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
                        <MenuItem value={ 0 }>Select an App</MenuItem>
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
                    { upload && (
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
                                <Button variant="outlined"
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

export default AppSelection;
