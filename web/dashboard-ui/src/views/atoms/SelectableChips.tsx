import { Check } from "@mui/icons-material";
import { Chip, ChipProps, ChipTypeMap } from "@mui/material";
import { forwardRef, useState } from "react";
import { useController, UseControllerProps } from "react-hook-form";

type toggleChipProps = { variant: ChipTypeMap['props']['variant'], onClick?: () => void } & ChipProps

export const FormSelectableChip = forwardRef((props: UseControllerProps<any> & ChipProps, ref) => {
    const { field } = useController(props);

    return <SelectableChip onChecked={field.onChange} onBlur={field.onBlur} {...props} />
})

export const SelectableChip: React.FC<{ checked?: boolean, onChecked?: (...event: any[]) => void } & ChipProps> = ({ onChecked, ...props }) => {
    const [mouseovered, setMouseovered] = useState(false);

    const [checked, setChecked] = useState<boolean>(props.defaultChecked || false);

    if (props.checked !== undefined && props.checked !== checked) {
        setChecked(props.checked);
    }

    const toggleChecked = () => { setChecked(!checked); onChecked && onChecked(!checked); }

    const unCheckedChipProps: toggleChipProps = { variant: "outlined", onClick: toggleChecked }
    const checkedChipProps: toggleChipProps = { variant: "filled", onClick: toggleChecked, onDelete: toggleChecked, deleteIcon: <Check /> }
    const checkedChipPropsMouseovered: toggleChipProps = { variant: "filled", onClick: toggleChecked, onDelete: toggleChecked }

    return <Chip {...(checked
        ? (mouseovered ? checkedChipPropsMouseovered : checkedChipProps)
        : unCheckedChipProps)}
        onMouseEnter={() => { setMouseovered(true) }}
        onMouseLeave={() => { setMouseovered(false) }}
        {...props} />
}
