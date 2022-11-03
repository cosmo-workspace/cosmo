// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';
import { createSerializer } from '@emotion/jest';

// https://github.com/mui-org/material-ui/issues/21701
// expect.addSnapshotSerializer({
// 	test: function (val) {
// 		return val && typeof val === "string" && val.indexOf("mui-") >= 0;
// 	},
// 	print: function (val) {
// 		let str = val as string;
// 		str = str.replace(/mui-[0-9]*/g, "mui-0000_for_toMatchSnapShot");
// 		return `"${str}"`;
// 	}
// });

expect.addSnapshotSerializer(createSerializer())

//globalThis.IS_REACT_ACT_ENVIRONMENT = true;