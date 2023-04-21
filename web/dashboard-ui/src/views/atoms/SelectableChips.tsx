import { CheckCircleOutline } from "@mui/icons-material";
import { Chip, ChipProps, ChipTypeMap } from "@mui/material";
import { ReactElement, useLayoutEffect, useRef, useState } from "react";
import { useController, UseControllerProps } from "react-hook-form";

type SelectableChipProps = { label: string, color?: ChipProps["color"], defaultChecked?: boolean } & UseControllerProps<any>;

type toggleChipProps = { variant: ChipTypeMap['props']['variant'], onClick?: () => void, onDelete?: () => void, deleteIcon?: ReactElement<any> }

export const SelectableChip: React.FC<SelectableChipProps> = (props) => {
    const { field } = useController(props);

    const [width, setWidth] = useState<number | undefined>(undefined);
    const ref = useRef<HTMLDivElement>(null);

    useLayoutEffect(() => {
        if (ref.current) {
            setWidth(ref.current.offsetWidth);
        }
    }, []);

    const [mouseovered, setMouseovered] = useState(false);

    const [checked, setChecked] = useState<boolean>(props.defaultChecked || false);
    const toggleChecked = () => { setChecked(!checked); field.onChange(!checked); }

    const unCheckedChipProps: toggleChipProps = { variant: "outlined", onClick: toggleChecked }
    const unCheckedChipPropsMouseovered: toggleChipProps = {
        variant: "outlined", onClick: toggleChecked, onDelete: toggleChecked, deleteIcon: <CheckCircleOutline />
    }
    const checkedChipProps: toggleChipProps = { variant: "filled", onClick: toggleChecked }
    const checkedChipPropsMouseovered: toggleChipProps = { variant: "filled", onClick: toggleChecked, onDelete: toggleChecked }

    return <Chip {...(checked
        ? (mouseovered ? checkedChipPropsMouseovered : checkedChipProps)
        : (mouseovered ? unCheckedChipPropsMouseovered : unCheckedChipProps))}
        onMouseEnter={() => { setMouseovered(true) }}
        onMouseLeave={() => { setMouseovered(false) }}
        color={props.color} label={props.label} onBlur={field.onBlur}
        ref={ref} sx={{ width: (width && width + 5), m: 0.05 }} />
}
