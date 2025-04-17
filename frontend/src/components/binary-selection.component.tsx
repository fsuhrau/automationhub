import React, {ChangeEvent, ReactElement, useCallback, useEffect, useState} from 'react';
import {
    Box,
    Input,
    LinearProgress,
    LinearProgressProps,
    MenuItem,
    SelectChangeEvent,
    Typography,
} from '@mui/material';
import Button from '@mui/material/Button';
import {getAppBundles, uploadNewApp} from '../services/app.service';
import Select from '@mui/material/Select';
import {IAppBinaryData} from "../types/app";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useDropzone} from "react-dropzone";
import {useApplicationContext} from "../hooks/ApplicationProvider";
import Grid from "@mui/material/Grid2";
import {useError} from "../ErrorProvider";

function LinearProgressWithLabel(props: LinearProgressProps & { value: number }): ReactElement {
    return (
        <Box sx={{display: 'flex', alignItems: 'center'}}>
            <Box sx={{width: '100%', mr: 1}}>
                <LinearProgress variant="determinate" {...props} />
            </Box>
            <Box sx={{minWidth: 35}}>
                <Typography variant="body2">{`${Math.round(
                    props.value,
                )}%`}</Typography>
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

    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {setError} = useError()

    const {binaryId, onSelectionChanged, upload} = props;

    const [state, setState] = useState<{
        binary: IAppBinaryData | null,
        binaries: IAppBinaryData[],
        selectedBinaryId: number | null,
        isUploading: boolean,
        uploadProgress: number,
        isProcessing: boolean
        isPreparing: boolean
    }>({
        binary: null,
        binaries: [],
        selectedBinaryId: binaryId,
        isUploading: false,
        uploadProgress: 0,
        isProcessing: false,
        isPreparing: false,
    })

    useEffect(() => {
        if (appId !== null) {
            getAppBundles(projectIdentifier, appId as number).then(response => {
                setState(prevState => ({...prevState, binaries: response.data}));
            }).catch(ex => {
                setError(ex)
            });
        }
    }, [binaryId, projectIdentifier, appId]);

    useEffect(() => {
        if (state.binary !== undefined && state.binary !== null && state.binary.ID !== binaryId) {
            onSelectionChanged(state.binary!);
        }
    }, [state.binary, binaryId, onSelectionChanged]);

    const selectUploadFile = (e: ChangeEvent<HTMLInputElement>): void => {
        e.preventDefault();
        if (e.target.files != null) {
            setState(prevState => ({...prevState, isProcessing: true, isPreparing: false}));
            const f = e.target.files[0] as File;
            setState(prevState => ({...prevState, isUploading: true}));
            uploadNewApp(f, projectIdentifier, appId as number, progress => {
                setState(prevState => ({...prevState, uploadProgress: progress, isUploading: progress < 100, isProcessing: progress === 100}));
            }, data => {
                if (data.data !== undefined && data.data !== null) {
                    setState(prevState => ({...prevState, uploadProgress: 100, isUploading: false, binary: data.data, selectedBinaryId: data.data.ID, binaries: [...prevState.binaries, data.data], isProcessing: false}));
                }
            });
        } else {
            setState(prevState => ({...prevState, isPreparing: false}));
        }
    };

    const handleChange = (e: SelectChangeEvent<number | null>): void => {
        e.preventDefault();
        if (e.target.value !== undefined) {
            const binaryId = e.target.value as number;
            const a = state.binaries.find(element => element.ID == binaryId);
            setState(prevState => ({...prevState, binary: a!, selectedBinaryId: binaryId}));
        }
    };

    const onDrop = useCallback((acceptedFiles: any) => {
        console.log(acceptedFiles)
        // Do something with the files
    }, [])
    const {getRootProps, getInputProps, isDragActive} = useDropzone({onDrop})


    const OnFileUploadClicked = () => {
        setState(prevState => ({...prevState, isPreparing: true}));
    }

    return (
        <Grid
            container={true}
            spacing={2}
            justifyContent="center"
            alignItems="center"
            direction={'column'}
        >
            <Grid size={12}>
                <Select
                    fullWidth={true}
                    id="app-selection"
                    value={state.selectedBinaryId}
                    onChange={event => handleChange(event)}
                    label={'App'}
                >
                    <MenuItem value={0}>Select an App</MenuItem>
                    {state.binaries.map((a) =>
                        <MenuItem key={`bin_select_item_${a.ID}`}
                                  value={a.ID}>{a.Platform} {a.Name} ({a.Version})</MenuItem>,
                    )}
                </Select>
            </Grid>
            <Grid size={12} container={true} justifyContent="center" alignItems="center">
                {upload && (
                    <Grid size={8}>
                        <LinearProgressWithLabel value={state.uploadProgress}/>
                    </Grid>
                )}
                {upload && (
                    <Grid size={4} justifyContent="center" alignItems="center" container={true} textAlign={"center"}
                          justifyItems={"center"} justifySelf={"center"}>
                        <label htmlFor="app-upload">
                            <Input
                                id="app-upload"
                                type="file"
                                sx={{visibility: 'hidden'}}
                                onClick={OnFileUploadClicked}
                                onChange={selectUploadFile}/>
                            { <Button variant="contained" component={"span"}>{state.isUploading ? "Uploading..." : state.isProcessing ? "Processing..." : state.isPreparing ? "Preparing..." : "Select File"}</Button> }
                        </label>
                    </Grid>
                )}
            </Grid>
        </Grid>
    );
};

export default BinarySelection;
