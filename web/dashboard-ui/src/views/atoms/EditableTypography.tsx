import { Check, Close, Edit } from "@mui/icons-material";
import {
    IconButton,
    InputBase, InputBaseProps,
    Stack,
    Tooltip
} from "@mui/material";
import React, { useState } from "react";

export type EditableTypographyProps =
    InputBaseProps
    & { children: string, onSave: (inputData: string) => void, showAlways?: boolean };

export const EditableTypography: React.FC<EditableTypographyProps> = ({ children, onSave, showAlways, ...props }) => {

    const [showEditIcon, setShowEditIcon] = useState(showAlways);
    const [editting, setEditing] = useState(false);
    const [inputData, setInputData] = useState<string>(children);

    const editIcon = (
        <Tooltip title="Edit" placement="top">
            <IconButton disableRipple size="small" type="button" sx={{ ml: 1 }}
                onClick={() => { setEditing(true) }}>
                <Edit />
            </IconButton>
        </Tooltip>
    );
    const editingIcons = (
        <Stack direction="row">
            <Tooltip title="Save" placement="top">
                <IconButton disableRipple size="small" type="button" sx={{ ml: 1 }}
                    onClick={() => { onSave(inputData); setEditing(false) }}>
                    <Check />
                </IconButton>
            </Tooltip>
            <Tooltip title="Cancel" placement="top">
                <IconButton disableRipple size="small" type="button" sx={{ ml: 1 }}
                    onClick={() => {
                        setInputData(children); setEditing(false)
                    }}>
                    <Close />
                </IconButton>
            </Tooltip>
        </Stack>
    );


    return (
        <InputBase
            size="small"
            fullWidth
            placeholder="Name"
            endAdornment={editting ? editingIcons : showEditIcon ? editIcon : undefined}
            defaultValue={children}
            readOnly={!editting}
            autoFocus={editting}
            onMouseOver={() => { setShowEditIcon(showAlways || true) }}
            onMouseLeave={() => { setShowEditIcon(showAlways || false) }}
            onBlur={(e) => { setInputData(e.currentTarget.value) }}
            sx={{ borderBottom: editting ? 1 : 0 }}
            {...props}
        />)
}
