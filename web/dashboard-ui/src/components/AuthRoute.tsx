import React, { ReactElement } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { isAdminUser } from '../views/organisms/UserModule';
import { useLogin } from './LoginProvider';

type Props = {
  children: ReactElement;
  admin?: boolean;
}

export const AuthRoute: React.VFC<Props> = ({ children, admin }) => {
  const { loginUser } = useLogin();
  const isAdmin = isAdminUser(loginUser);
  let location = useLocation();

  if (!loginUser) {
    return (<Navigate
      to={{ pathname: '/signin', search: `from=${location.pathname}` }}
      state={{ from: location }}
      replace
    />);
  } else if (admin && !isAdmin) {
    return (<Navigate to='/' replace />);
  } else {
    return children;
  }
}