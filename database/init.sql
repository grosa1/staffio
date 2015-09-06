-- init

INSERT INTO oauth_scope(name,label,description,is_default) VALUES('basic', 'Basic', 'Read your Uid (login name) and Nickname', true);
INSERT INTO oauth_scope(name,label,description) VALUES('profile', 'Personal Information', 'Read your GivenName, Surname, Email, etc.');

-- INSERT INTO oauth_client VALUES(1, '1234', 'Demo', 'aabbccdd', 'http://localhost:3000/appauth', '{}', now());
