-- ============================================================
-- ComplianceForge — Seed: NIST CSF 2.0 Subcategories
-- ============================================================

-- GOVERN (GV)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.OC-01','Organisational context is understood','directive','management',1),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.OC-02','Internal and external stakeholders are understood','directive','management',2),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.OC-03','Legal, regulatory and contractual requirements are understood','directive','management',3),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.OC-04','Critical objectives, capabilities and services are understood','directive','management',4),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.OC-05','Outcomes, capabilities and services that depend on supply chain are understood','directive','management',5),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RM-01','Risk management objectives are established and agreed','directive','management',6),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RM-02','Risk appetite and tolerance are established','directive','management',7),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RM-03','Activities and outcomes of cybersecurity risk management are included in ERM','directive','management',8),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RM-04','Strategic direction describing appropriate risk response options is established','directive','management',9),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RR-01','Organisational leadership accepts cybersecurity risk','directive','management',10),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RR-02','Roles and responsibilities for cybersecurity risk management are established','directive','administrative',11),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RR-03','Adequate resources are allocated for cybersecurity','directive','management',12),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.RR-04','Cybersecurity is included in human resources practices','directive','administrative',13),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.PO-01','Cybersecurity risk management policy is established','directive','administrative',14),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.PO-02','Cybersecurity risk management policy is communicated and enforced','directive','administrative',15),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000001','GV.SC-01','Cybersecurity supply chain risk management is performed','directive','management',16);

-- IDENTIFY (ID)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.AM-01','Physical devices and systems are inventoried','preventive','administrative',17),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.AM-02','Software platforms and applications are inventoried','preventive','administrative',18),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.AM-03','Data flows are mapped','preventive','administrative',19),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.AM-04','External information systems are catalogued','preventive','administrative',20),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.AM-05','Assets are prioritised based on classification and criticality','preventive','administrative',21),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.RA-01','Vulnerabilities in assets are identified, validated and recorded','detective','technical',22),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.RA-02','Cyber threat intelligence is received from information sharing forums','detective','administrative',23),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.RA-03','Internal and external threats are identified and recorded','detective','administrative',24),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.RA-04','Potential business impacts and likelihoods are identified','detective','administrative',25),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.RA-05','Threats, vulnerabilities, likelihoods and impacts are used to determine risk','detective','administrative',26),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.RA-06','Risk responses are identified and prioritised','corrective','administrative',27),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.IM-01','Improvements are identified from evaluations','corrective','administrative',28),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.IM-02','Improvements are identified from security tests and exercises','corrective','administrative',29),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000002','ID.IM-03','Improvements are identified from operational incidents','corrective','administrative',30);

-- PROTECT (PR) — Key subcategories
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AA-01','Identities and credentials for authorised users are managed','preventive','technical',31),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AA-02','Identities are proofed and bound to credentials','preventive','technical',32),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AA-03','Users, services and hardware are authenticated','preventive','technical',33),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AA-04','Identity assertions are protected, conveyed and verified','preventive','technical',34),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AA-05','Access permissions, entitlements and authorisations are managed','preventive','technical',35),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AA-06','Physical access to assets is managed, monitored and enforced','preventive','physical',36),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AT-01','Personnel are provided cybersecurity awareness and training','preventive','administrative',37),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.AT-02','Privileged users understand roles and responsibilities','preventive','administrative',38),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.DS-01','Data-at-rest is protected','preventive','technical',39),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.DS-02','Data-in-transit is protected','preventive','technical',40),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.DS-10','Data in use is protected','preventive','technical',41),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.DS-11','Backups of data are maintained and tested','preventive','technical',42),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.PS-01','Configuration management practices are established','preventive','technical',43),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.PS-02','Software is maintained, replaced and removed','preventive','technical',44),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.PS-03','Hardware is maintained, replaced and removed','preventive','physical',45),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.PS-04','Log records are generated and made available','detective','technical',46),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.PS-05','Installation of software is prevented or restricted','preventive','technical',47),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.PS-06','Secure software development practices are integrated','preventive','technical',48),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.IR-01','Networks and environments are protected from unauthorised logical access','preventive','technical',49),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.IR-02','Network integrity is protected incorporating network segregation','preventive','technical',50),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.IR-03','Mechanisms to achieve resilience requirements in normal and adverse situations','preventive','technical',51),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000003','PR.IR-04','Adequate resource capacity to ensure availability is maintained','preventive','technical',52);

-- DETECT (DE)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.CM-01','Networks and network services are monitored to find potentially adverse events','detective','technical',53),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.CM-02','The physical environment is monitored to find potentially adverse events','detective','physical',54),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.CM-03','Personnel activity is monitored to find potentially adverse events','detective','technical',55),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.CM-06','External service provider activities are monitored','detective','administrative',56),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.CM-09','Computing hardware and software are monitored to find potentially adverse events','detective','technical',57),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.AE-02','Potentially adverse events are analysed to better understand associated activities','detective','technical',58),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.AE-03','Information is correlated from multiple sources','detective','technical',59),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000004','DE.AE-06','Information on adverse events is provided to authorised staff','detective','administrative',60);

-- RESPOND (RS)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.MA-01','The incident response plan is executed in coordination with relevant third parties','corrective','administrative',61),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.MA-02','Incident reports are triaged and validated','corrective','administrative',62),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.MA-03','Incidents are categorised and prioritised','corrective','administrative',63),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.MA-04','Incidents are escalated or elevated as needed','corrective','administrative',64),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.AN-03','Analysis is performed to determine what has taken place during an incident','detective','technical',65),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.AN-06','Actions performed during an investigation are recorded','detective','administrative',66),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.AN-07','Incident data and metadata are collected and integrity is preserved','detective','technical',67),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.AN-08','An incidents magnitude is estimated and validated','detective','administrative',68),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.CO-02','Internal and external stakeholders are notified of incidents','corrective','administrative',69),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.CO-03','Information is shared with designated internal and external stakeholders','corrective','administrative',70),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.MI-01','Incidents are contained','corrective','technical',71),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000005','RS.MI-02','Incidents are eradicated','corrective','technical',72);

-- RECOVER (RC)
INSERT INTO framework_controls (framework_id, domain_id, code, title, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000006','RC.RP-01','Recovery portion of the incident response plan is executed','corrective','administrative',73),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000006','RC.RP-02','Recovery actions are selected, scoped, prioritised and performed','corrective','technical',74),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000006','RC.RP-03','Integrity of backups and other restoration assets is verified','detective','technical',75),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000006','RC.RP-04','Critical functions are recovered','corrective','technical',76),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000006','RC.RP-05','Integrity of restored assets is verified','detective','technical',77),
('a0000001-0000-0000-0000-000000000006','d0000006-0001-0000-0000-000000000006','RC.CO-03','Recovery activities and progress are communicated to stakeholders','corrective','administrative',78);

-- ============================================================
-- Cyber Essentials — 5 Technical Controls (expanded)
-- ============================================================
INSERT INTO framework_controls (framework_id, domain_id, code, title, description, control_type, implementation_type, sort_order) VALUES
('a0000001-0000-0000-0000-000000000004','d0000004-0001-0000-0000-000000000001','CE-FW-01','Firewalls and Internet Gateways','Ensure boundary firewalls and internet gateways are configured to protect against unauthorised access from the internet. All devices that connect to the internet must be protected by a correctly configured firewall.','preventive','technical',1),
('a0000001-0000-0000-0000-000000000004','d0000004-0001-0000-0000-000000000002','CE-SC-01','Secure Configuration','Computers and network devices should be configured to reduce the level of inherent vulnerabilities and provide only the services required to fulfil their role. Remove or disable unnecessary software, accounts, and services.','preventive','technical',2),
('a0000001-0000-0000-0000-000000000004','d0000004-0001-0000-0000-000000000003','CE-SU-01','Security Update Management','Ensure that devices and software are not vulnerable to known security issues for which fixes are available. Keep all software, operating systems and firmware up-to-date, applying patches within 14 days of release.','preventive','technical',3),
('a0000001-0000-0000-0000-000000000004','d0000004-0001-0000-0000-000000000004','CE-AC-01','User Access Control','Ensure that only authorised individuals have user accounts with access to systems and data. Manage user accounts including privileged accounts. Implement MFA where available.','preventive','technical',4),
('a0000001-0000-0000-0000-000000000004','d0000004-0001-0000-0000-000000000005','CE-MP-01','Malware Protection','Restrict execution of known malware and untrusted software to prevent harmful code from causing damage or accessing sensitive data. Use anti-malware software or application whitelisting.','preventive','technical',5);
