import React, {useEffect, useState} from "react";
import {Dialog, DialogActions, DialogContent, TextField, Typography} from "@mui/material";
import Button from "@mui/material/Button";

export type EditAttribute = {
    attribute: string | null,
    value: string,
}

interface EditAttributePopupProps {
    attribute: string | null,
    value: string,
    onClose: () => void,
    onSubmit: (attribute: string, value: string) => void,
}

const EditAttributePopup: React.FC<EditAttributePopupProps> = (props: EditAttributePopupProps) => {

    const [state, setState] = useState<{
        value: string,
    }>({
        value: props.value,
    });

    const open = props.attribute !== null && props.attribute !== '';

    const handleEditClose = () => {
        props.onClose();
    };

    const handleEditSubmit = () => {
        props.onSubmit(props.attribute!, state.value!)
        handleEditClose();
    };

    useEffect(() => {
        setState(prevState => ({...prevState, value: props.value}))
    }, [props.attribute, props.value]);

    return (
        <Dialog open={open} onClose={handleEditClose}>
            <DialogContent>
                <Typography variant={"body1"}>{props.attribute}: {props.value}</Typography>
                <TextField
                    autoFocus
                    margin="dense"
                    id="edit_attribute"
                    placeholder={props.attribute ? props.attribute : 'Attribute'}
                    value={state.value}
                    onChange={(e) => setState(prevState => ({
                        ...prevState,
                        value: e.target.value
                    }))}
                    fullWidth
                    variant="standard"
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleEditClose}>Cancel</Button>
                <Button onClick={handleEditSubmit}>Save</Button>
            </DialogActions>
        </Dialog>
    )
}

export default EditAttributePopup;