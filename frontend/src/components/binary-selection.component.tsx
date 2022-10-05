import React, { ChangeEvent, ReactElement, useCallback, useEffect, useState } from 'react';
import Grid from '@mui/material/Grid';
import {
    Box,
    Button,
    FormControl,
    Input,
    InputLabel,
    LinearProgress,
    LinearProgressProps,
    MenuItem,
    SelectChangeEvent,
    Typography,
} from '@mui/material';
import { getAllApps, getAppBundles, updateAppBundle, uploadNewApp } from '../services/app.service';
import Select from '@mui/material/Select';
import { IAppBinaryData } from "../types/app";
import { useParams } from "react-router-dom";
import { useProjectAppContext } from "../project/app.context";
import { useDropzone } from "react-dropzone";

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


interface BinarySelectionProps {
    binaryId: number | null,
    onSelectionChanged: (app: IAppBinaryData) => void;
    upload: boolean;
}

const BinarySelection: React.FC<BinarySelectionProps> = (props) => {
    let params = useParams();

    const { projectId, appId } = useProjectAppContext();

    const { binaryId, onSelectionChanged, upload } = props;

    const [binary, setBinary] = useState<IAppBinaryData>();
    const [binaries, setBinaries] = useState<IAppBinaryData[]>([]);
    const [selectedAppID, setSelectedAppID] = useState<number>(1);

    const [uploadProgress, setUploadProgress] = useState<number>(0);

    useEffect(() => {
        if (appId !== null) {
            getAppBundles(params.project_id as string, appId).then(response => {
                setBinaries(response.data);
            }).catch(e => {
            });
        }
    }, [binaryId, params.project_id]);

    useEffect(() => {
        if (binary !== undefined && binary.ID !== binaryId) {
            onSelectionChanged(binary);
        }
    }, [binary, binaryId, onSelectionChanged]);

    const selectUploadFile = (e: ChangeEvent<HTMLInputElement>): void => {
        e.preventDefault();
        if (e.target.files != null) {
            const f = e.target.files[ 0 ] as File;
            uploadNewApp(f, projectId, appId, progress => {
                setUploadProgress(progress);
            }, data => {
                if (data.data !== undefined && data.data !== null) {
                    setBinaries(prevState => {
                        const newState = [...prevState];
                        newState.push(data.data);
                        return newState;
                    });
                    setBinary(data.data);
                    setSelectedAppID(data.data.ID);
                }
            });
        }
    };

    const handleChange = (e: SelectChangeEvent<number>): void => {
        e.preventDefault();
        if (e.target.value !== undefined) {
            const appId = e.target.value as number;
            const a = binaries.find(element => element.ID == appId);
            setBinary(a);
            setSelectedAppID(appId);
        }
    };

    const onDrop = useCallback((acceptedFiles: any) => {
        console.log(acceptedFiles)
        // Do something with the files
    }, [])
    const {getRootProps, getInputProps, isDragActive} = useDropzone({onDrop})

    return (
        <Grid
            container={ true }
            spacing={ 2 }
            justifyContent="center"
            alignItems="center"
            direction={ 'column' }
        >
            <Grid item={ true }>
                <FormControl>
                    <InputLabel id="app-selection-label">App</InputLabel>
                    <Select
                        labelId="app-selection-label"
                        id="app-selection"
                        value={ selectedAppID }
                        onChange={ event => handleChange(event) }
                        label={ 'App' }
                    >
                        <MenuItem value={ 0 }>Select an App</MenuItem>
                        { binaries.map((a) =>
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
                >
                    { upload && (
                        <Grid item={ true }>
                            <Box sx={ { width: '250px' } }>
                                <LinearProgressWithLabel value={ uploadProgress }/>
                            </Box>
                        </Grid>
                    ) }
                    { upload && (
                        <Grid item={ true }>
                            <label htmlFor="app-upload">
                                <Input
                                    id="app-upload"
                                    type="file"
                                    sx={{visibility: 'hidden'}}
                                    onChange={ selectUploadFile }/>
                                <Button variant="contained" component="span">
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

export default BinarySelection;
