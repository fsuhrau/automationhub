import React, {ChangeEvent, ReactElement, useCallback, useEffect, useState} from 'react';
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
import {getAppBundles, uploadNewApp} from '../services/app.service';
import Select from '@mui/material/Select';
import {IAppBinaryData} from "../types/app";
import {useParams} from "react-router-dom";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useDropzone} from "react-dropzone";
import {useApplicationContext} from "../hooks/ApplicationProvider";
import Grid from "@mui/material/Grid2";

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

    const {binaryId, onSelectionChanged, upload} = props;

    const [binary, setBinary] = useState<IAppBinaryData>();
    const [binaries, setBinaries] = useState<IAppBinaryData[]>([]);
    const [selectedBinaryID, setSelectedBinaryID] = useState<number | null>(binaryId);

    const [uploadProgress, setUploadProgress] = useState<number>(0);

    useEffect(() => {
        if (appId !== null) {
            getAppBundles(projectIdentifier, appId as number).then(response => {
                setBinaries(response.data);
            }).catch(e => {
            });
        }
    }, [binaryId, projectIdentifier, appId]);

    useEffect(() => {
        if (binary !== undefined && binary.ID !== binaryId) {
            onSelectionChanged(binary);
        }
    }, [binary, binaryId, onSelectionChanged]);

    const selectUploadFile = (e: ChangeEvent<HTMLInputElement>): void => {
        e.preventDefault();
        if (e.target.files != null) {
            const f = e.target.files[0] as File;
            uploadNewApp(f, projectIdentifier, appId  as number, progress => {
                setUploadProgress(progress);
            }, data => {
                if (data.data !== undefined && data.data !== null) {
                    setBinaries(prevState => {
                        const newState = [...prevState];
                        newState.push(data.data);
                        return newState;
                    });
                    setBinary(data.data);
                    setSelectedBinaryID(data.data.ID);
                }
            });
        }
    };

    const handleChange = (e: SelectChangeEvent<number|null>): void => {
        e.preventDefault();
        if (e.target.value !== undefined) {
            const binaryId = e.target.value as number;
            const a = binaries.find(element => element.ID == binaryId);
            setBinary(a);
            setSelectedBinaryID(binaryId);
        }
    };

    const onDrop = useCallback((acceptedFiles: any) => {
        console.log(acceptedFiles)
        // Do something with the files
    }, [])
    const {getRootProps, getInputProps, isDragActive} = useDropzone({onDrop})

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
                    value={selectedBinaryID}
                    onChange={event => handleChange(event)}
                    label={'App'}
                >
                    <MenuItem value={0}>Select an App</MenuItem>
                    {binaries.map((a) =>
                        <MenuItem key={`bin_select_item_${a.ID}`}
                                  value={a.ID}>{a.Platform} {a.Name} ({a.Version})</MenuItem>,
                    )}
                </Select>
            </Grid>
            <Grid size={12} container={true} justifyContent="center" alignItems="center">
                {upload && (
                    <Grid size={8}>
                            <LinearProgressWithLabel value={uploadProgress}/>
                    </Grid>
                )}
                {upload && (
                    <Grid size={4} justifyContent="center" alignItems="center" container={true} textAlign={"center"} justifyItems={"center"} justifySelf={"center"}>
                        <label htmlFor="app-upload">
                            <Input
                                id="app-upload"
                                type="file"
                                sx={{visibility: 'hidden'}}
                                onChange={selectUploadFile}/>
                            <Button variant="contained" component="span">
                                Upload New
                            </Button>
                        </label>
                    </Grid>
                )}
            </Grid>
        </Grid>
    );
};

export default BinarySelection;
