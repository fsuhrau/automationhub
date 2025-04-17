import React, { useState } from 'react';
import Grid from '@mui/material/Grid2';
import {
    Button,
    Dialog,
    IconButton,
    Slide,
    TextField,
    Typography
} from '@mui/material';
import { TransitionProps } from "@mui/material/transitions";
import { Close } from "@mui/icons-material";
import IProject from "./project";

const Transition = React.forwardRef(function Transition(
    props: TransitionProps & {
        children: React.ReactElement;
    },
    ref: React.Ref<unknown>,
) {
    return <Slide direction="left" ref={ ref } { ...props } />;
});

export interface CreateProjectDialogProps {
    open: boolean;
    onSubmit: (data: IProject) => void;
    onClose: () => void;
}

const CreateProjectDialog: React.FC<CreateProjectDialogProps> = (props: CreateProjectDialogProps) => {

    const {open, onClose, onSubmit} = props;

    const [state, setState] = useState<IProject>({
        Name: '',
        Identifier: '',
    } as IProject)

    const handleClose = () => {
        onClose();
    };

    const submit = () => {
        onSubmit(state);
        handleClose();
    }

    return (
        <Dialog
            fullScreen
            open={ open }
            onClose={ handleClose }
            TransitionComponent={ Transition }
        >
            <Grid container={ true } sx={ {padding: 5} }>
                <Grid size={12}>
                    <Typography variant={ "h5" }><IconButton onClick={ onClose }><Close/></IconButton>Create a new
                        Project</Typography>
                </Grid>
                <Grid size={2} container={ true } sx={ {padding: 3} } spacing={ 2 }>
                    <Grid size={ 12 }>
                        <Typography variant={ "caption" }>Project Name</Typography>
                    </Grid>
                    <Grid size={ 12 }>
                        <TextField
                            hiddenLabel
                            id="project_name"
                            variant="filled"
                            value={ state.Name }
                            size="small"
                            fullWidth={true}
                            onChange={ event => setState(prevState => ({...state, Name: event.target.value})) }
                        />
                    </Grid>
                    <Grid size={ 12 }>
                        <Button variant={"outlined"} onClick={submit}>Create</Button>
                    </Grid>
                </Grid>
            </Grid>
        </Dialog>
    );
};

export default CreateProjectDialog;
