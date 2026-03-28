-- ============================================================
-- ComplianceForge — Seed: PCI DSS v4.0 Controls
-- Key controls across all 12 requirements
-- ============================================================

-- Req 1: Network Security Controls
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.1.1','Processes and mechanisms for network security controls are defined and understood','directive','administrative',1),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.2.1','Inbound and outbound traffic is restricted to that which is necessary','preventive','technical',2),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.2.5','All services protocols and ports allowed are identified and approved','preventive','technical',3),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.2.8','NSCs are configured between wireless and CDE networks','preventive','technical',4),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.3.1','Inbound traffic to the CDE is restricted','preventive','technical',5),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.3.2','Outbound traffic from the CDE is restricted','preventive','technical',6),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.4.1','NSCs are implemented between trusted and untrusted networks','preventive','technical',7),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.4.2','Inbound traffic from untrusted to trusted networks is restricted','preventive','technical',8),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000001','1.5.1','Security controls on computing devices connecting to untrusted networks','preventive','technical',9);

-- Req 2: Secure Configurations
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000002','2.1.1','Processes for applying secure configurations are defined and understood','directive','administrative',10),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000002','2.2.1','Configuration standards are developed maintained and implemented','preventive','technical',11),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000002','2.2.2','Vendor default accounts are managed','preventive','technical',12),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000002','2.2.4','Only necessary services and protocols are enabled','preventive','technical',13),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000002','2.2.5','Insecure services and protocols are managed','preventive','technical',14),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000002','2.2.7','All non-console administrative access is encrypted with strong cryptography','preventive','technical',15);

-- Req 3: Protect Stored Account Data
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000003','3.1.1','Processes for protecting stored account data are defined and understood','directive','administrative',16),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000003','3.2.1','Account data storage is kept to a minimum','preventive','administrative',17),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000003','3.3.1','SAD is not stored after authorisation','preventive','technical',18),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000003','3.4.1','PAN is masked when displayed','preventive','technical',19),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000003','3.5.1','PAN is rendered unreadable anywhere it is stored','preventive','technical',20),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000003','3.6.1','Cryptographic keys used to protect stored account data are protected','preventive','technical',21),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000003','3.7.1','Key management policies and procedures are implemented','directive','administrative',22);

-- Req 4: Protect Data in Transit
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000004','4.1.1','Processes for protecting cardholder data in transit are defined','directive','administrative',23),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000004','4.2.1','PAN is protected with strong cryptography during transmission over open networks','preventive','technical',24),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000004','4.2.1.1','An inventory of trusted keys and certificates is maintained','preventive','technical',25),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000004','4.2.2','PAN is secured with strong cryptography when sent via end-user messaging','preventive','technical',26);

-- Req 5: Malware Protection
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.1.1','Processes for protecting against malware are defined and understood','directive','administrative',27),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.2.1','Anti-malware solutions are deployed on all system components','preventive','technical',28),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.2.2','Anti-malware solutions detect all known types of malware','detective','technical',29),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.2.3','System components not at risk for malware are evaluated periodically','detective','administrative',30),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.3.1','Anti-malware mechanisms and definitions are kept current','preventive','technical',31),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.3.2','Anti-malware performs periodic scans and active or real-time scans','detective','technical',32),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.3.3','Anti-malware for removable electronic media performs scans automatically','detective','technical',33),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.3.4','Audit logs for anti-malware are enabled and retained','detective','technical',34),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000005','5.4.1','Mechanisms detect and protect against phishing attacks','preventive','technical',35);

-- Req 6: Secure Systems and Software
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000006','6.1.1','Processes for developing secure systems and software are defined','directive','administrative',36),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000006','6.2.1','Custom software is developed securely','preventive','technical',37),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000006','6.2.4','Software engineering techniques prevent or mitigate common attacks','preventive','technical',38),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000006','6.3.1','Security vulnerabilities are identified and addressed','corrective','technical',39),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000006','6.3.3','Critical security patches are installed within one month of release','corrective','technical',40),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000006','6.4.1','Public-facing web applications are protected against attacks','preventive','technical',41),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000006','6.5.1','Changes to system components are managed via change control procedures','preventive','administrative',42);

-- Req 7: Restrict Access
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000007','7.1.1','Processes for restricting access are defined and understood','directive','administrative',43),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000007','7.2.1','Access control model is defined with appropriate assignments','preventive','administrative',44),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000007','7.2.2','Access is assigned based on job classification and function','preventive','administrative',45),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000007','7.2.4','All user accounts and related access privileges are reviewed','detective','administrative',46),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000007','7.2.5','All access by application and system accounts is appropriately assigned','preventive','technical',47),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000007','7.2.6','All access by application and system accounts is appropriately managed','preventive','technical',48);

-- Req 8: Identify and Authenticate
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.1.1','Processes for identification and authentication are defined','directive','administrative',49),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.2.1','All users are assigned a unique ID before access is allowed','preventive','technical',50),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.2.2','Group shared or generic accounts are not used','preventive','administrative',51),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.3.1','All user access to system components is authenticated via at least one factor','preventive','technical',52),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.3.2','Strong cryptography renders all authentication factors unreadable during storage','preventive','technical',53),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.3.6','Passwords/passphrases meet minimum complexity requirements','preventive','technical',54),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.3.9','Passwords/passphrases are changed at least once every 90 days','preventive','administrative',55),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.4.1','MFA is implemented for all non-console access into the CDE','preventive','technical',56),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.4.2','MFA is implemented for all access into the CDE','preventive','technical',57),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.5.1','MFA systems are implemented as follows and are not susceptible to replay','preventive','technical',58),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000008','8.6.1','Interactive login to system components is restricted','preventive','technical',59);

-- Req 9: Physical Access
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000009','9.1.1','Processes for restricting physical access are defined','directive','administrative',60),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000009','9.2.1','Physical access controls manage entry to facilities','preventive','physical',61),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000009','9.2.4','Physical access to sensitive areas is controlled','preventive','physical',62),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000009','9.4.1','All media with cardholder data is physically secured','preventive','physical',63),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000009','9.4.6','Hard-copy materials with cardholder data are destroyed when no longer needed','preventive','physical',64);

-- Req 10: Log and Monitor
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.1.1','Processes for logging and monitoring are defined','directive','administrative',65),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.2.1','Audit logs are enabled and active for all system components and cardholder data','detective','technical',66),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.2.2','Audit logs record sufficient detail for all auditable events','detective','technical',67),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.3.1','Read access to audit log files is limited','preventive','technical',68),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.3.3','Audit log files are promptly backed up','preventive','technical',69),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.4.1','Audit logs are reviewed at least once daily','detective','administrative',70),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.5.1','Audit log history is retained for at least 12 months','preventive','administrative',71),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.6.1','System clocks and time are synchronised using time-synchronisation technology','preventive','technical',72),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000010','10.7.1','Failures of critical security control systems are detected and addressed','detective','technical',73);

-- Req 11: Test Security
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.1.1','Processes for security testing are defined and understood','directive','administrative',74),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.2.1','Authorised and unauthorised wireless access points are managed','detective','technical',75),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.3.1','Internal vulnerability scans are performed at least every three months','detective','technical',76),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.3.2','External vulnerability scans are performed at least every three months','detective','technical',77),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.4.1','External penetration testing is performed at least every 12 months','detective','technical',78),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.5.1','Intrusion-detection or prevention techniques detect and alert on intrusions','detective','technical',79),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.5.2','A change-detection mechanism is deployed to alert on unauthorised modification','detective','technical',80),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000011','11.6.1','A change and tamper detection mechanism is deployed on payment pages','detective','technical',81);

-- Req 12: InfoSec Policies and Programs
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.1.1','An information security policy is established and published','directive','administrative',82),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.1.2','Information security policy is reviewed at least once every 12 months','detective','administrative',83),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.3.1','Each PCI DSS requirement that provides flexibility is addressed with a targeted risk analysis','detective','administrative',84),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.4.1','Compliance with the PCI DSS is managed','detective','administrative',85),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.5.2','PCI DSS scope is documented and validated at least every 12 months','detective','administrative',86),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.6.1','A formal security awareness program is implemented','preventive','administrative',87),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.6.2','Security awareness program is reviewed at least once every 12 months','detective','administrative',88),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.8.1','A list of all third-party service providers is maintained','preventive','administrative',89),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.8.2','Written agreements with TPSPs include acknowledgement of responsibility','preventive','administrative',90),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.8.5','Information about PCI DSS requirements managed by each TPSP is maintained','preventive','administrative',91),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.10.1','An incident response plan exists and is ready to be activated','corrective','administrative',92),
('a0000001-0000-0000-0000-000000000007','d0000007-0001-0000-0000-000000000012','12.10.2','The incident response plan is reviewed and tested at least every 12 months','detective','administrative',93);
