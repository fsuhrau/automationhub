import React, { useEffect, useState } from 'react';

import { Button, Card, CardMedia, Popover, Popper, Typography } from '@mui/material';
import Paper from "@mui/material/Paper";

interface CellExpandProps {
    id: number
    value: string
    data: string
}

const CellExpand: React.FC<CellExpandProps> = (props: CellExpandProps) => {
    const {id, value, data} = props;

    const [anchorLogScreenEl, setAnchorLogScreenEl] = useState<HTMLButtonElement | null>(null);
    const wrapper = React.useRef<HTMLDivElement | null>(null);
    const cellDiv = React.useRef(null);
    const cellValue = React.useRef(null);
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const [showFullCell, setShowFullCell] = useState(false);


    const showLogScreenPopup = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorLogScreenEl(event.currentTarget);
    };

    const hideLogScreenPopup = (): void => {
        setAnchorLogScreenEl(null);
    };

    const logScreenOpen = Boolean(anchorLogScreenEl);
    const logScreenID = logScreenOpen ? 'simple-popover' : undefined;


    const handleMouseEnter = () => {
        setAnchorEl(cellDiv.current);
        setShowFullCell(true);
    };

    const handleMouseLeave = () => {
        setShowFullCell(false);
    };

    useEffect(() => {
        if (!showFullCell) {
            return undefined;
        }

        function handleKeyDown(nativeEvent: KeyboardEvent) {
            // IE11, Edge (prior to using Bink?) use 'Esc'
            if (nativeEvent.key === "Escape" || nativeEvent.key === "Esc") {
                setShowFullCell(false);
            }
        }

        document.addEventListener("keydown", handleKeyDown);

        return () => {
            document.removeEventListener("keydown", handleKeyDown);
        };
    }, [setShowFullCell, showFullCell]);

    return (
        <div
            ref={ wrapper }
            onMouseEnter={ handleMouseEnter }
            onMouseLeave={ handleMouseLeave }
            style={{ display: "block", maxWidth: "inherit" }}
        >
            <div
                ref={ cellDiv }
                style={ {
                    height: 1,
                    display: "block",
                    top: 0
                } }
            />
            { value === '' && (<>
                <Button aria-describedby={ `$id` } variant="contained"
                        onClick={ showLogScreenPopup }>
                    Show
                </Button>
                <Popover
                    id={ logScreenID }
                    open={ logScreenOpen }
                    anchorEl={ anchorLogScreenEl }
                    onClose={ hideLogScreenPopup }
                    anchorOrigin={ {
                        vertical: 'bottom',
                        horizontal: 'left',
                    } }
                >
                    <Card>
                        <CardMedia
                            component="img"
                            height="400"
                            image={ `/api/data/${ data }` }
                            alt="green iguana"
                        />
                    </Card>
                </Popover>
            </>) }
            { value !== '' && (<div style={{ display: "block", maxWidth: "inherit" }}>
                <div ref={ cellValue } style={{ display: "block", maxWidth: "inherit", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{ value }</div>
                <Popper
                    open={ showFullCell && anchorEl != null }
                    anchorEl={ anchorEl }
                    style={{ maxWidth: '500px', marginLeft: -17 }}
                >
                    <Paper
                        elevation={1}>
                        <Typography variant={ "body2" } style={{ padding: 8 }}>{ value }</Typography>
                    </Paper>
                </Popper>
            </div>) }
        </div>
    );
};

export default CellExpand;
