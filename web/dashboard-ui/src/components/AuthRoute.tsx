import React from 'react';
import { Redirect, Route, RouteProps } from 'react-router-dom';
import { useLogin } from './LoginProvider';

type Props = RouteProps & {
  component: any;
  admin?: boolean;
}

export const AuthRoute = ({ component: Component, admin, ...rest }: Props) => {
  const { loginUser } = useLogin();
  const isAdmin = loginUser && loginUser.role === 'cosmo-admin';

  return (
    <Route
      {...rest}
      render={(props) => {
        console.log('props', props);
        if (!loginUser) {
          return (<Redirect
            to={{
              pathname: '/signin',
              search: `from=${props.location.pathname}`,
              state: { from: props.location },
            }}
          />);
        }
        if (admin && !isAdmin) {
          return (<Redirect to='/' />);
        }
        return (<Component {...props} />);
      }}
    />
  )
}
