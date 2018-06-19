package deploy

import (
	"encoding/json"
	"fmt"
	"lorhammer/src/tools"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

const typeAmazon = "amazon"

var logAmazon = logrus.WithField("logger", "orchestrator/deploy/amazon")

type amazonImpl struct {
	Region           string          `json:"region"`
	ImageID          string          `json:"imageId"`
	InstanceType     string          `json:"instanceType"`
	KeyPairName      string          `json:"keyPairName"`
	SecurityGroupIds []string        `json:"securityGroupIds"`
	MinCount         int64           `json:"minCount"`
	MaxCount         int64           `json:"maxCount"`
	DistantConfig    json.RawMessage `json:"distantConfig"`

	ec2Client       *ec2.EC2
	distantDeployer *amazonDistantImpl
	mqttAddress     string
	instancesID     []*string
}

type amazonDistantImpl struct {
	SSHKeyPath        string `json:"sshKeyPath"`
	User              string `json:"user"`
	IPServer          string `json:"ipServer"`
	PathFile          string `json:"pathFile"`
	PathWhereScp      string `json:"pathWhereScp"`
	BeforeCmd         string `json:"beforeCmd"`
	AfterCmd          string `json:"afterCmd"`
	NbDistantToLaunch int    `json:"nbDistantToLaunch"`
	cmdFabric         func(name string, arg ...string) *exec.Cmd
}

func newAmazonFromJSON(serialized json.RawMessage, mqttClient tools.Mqtt) (deployer, error) {
	client := &amazonImpl{}
	if err := json.Unmarshal(serialized, client); err != nil {
		return nil, err
	}

	b := true
	s, err := session.NewSession(&aws.Config{Region: aws.String(client.Region), CredentialsChainVerboseErrors: &b})
	if err != nil {
		return nil, err
	}

	client.ec2Client = ec2.New(s)
	if client.distantDeployer, err = newAmazonDistantImpl(client.DistantConfig); err != nil {
		return nil, err
	}
	client.mqttAddress = mqttClient.GetAddress()
	return client, nil
}

func newAmazonDistantImpl(serialized json.RawMessage) (*amazonDistantImpl, error) {
	d := &amazonDistantImpl{}
	if err := json.Unmarshal(serialized, d); err != nil {
		return nil, err
	}
	d.cmdFabric = exec.Command
	return d, nil
}

func (distant *amazonDistantImpl) Deploy() error {

	c := fmt.Sprintf("%s@%s:%s", distant.User, distant.IPServer, distant.PathWhereScp)
	logDistant.WithField("cmd", "scp -q -i "+distant.SSHKeyPath+" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+distant.PathFile+" "+c).Info("Will exec cmd")
	cmd := distant.cmdFabric("scp", "-q", "-i", distant.SSHKeyPath, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", distant.PathFile, c)
	if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
		logDistant.WithField("output", fmt.Sprintf("%s", stdoutStderr)).Info("Scp output")
		return err
	}
	return nil
}

func (distant *amazonDistantImpl) runCmd(cmd string) error {
	errs := distantRunError{Errors: make([]error, 0)}
	ip := fmt.Sprintf("%s@%s", distant.User, distant.IPServer)
	logDistant.WithField("cmd", "ssh -q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+ip+" "+cmd).Info("Will exec cmd")

	chanErr := make(chan error)
	defer close(chanErr)

	for i := 0; i < distant.NbDistantToLaunch; i++ {
		go func() {
			cmd := distant.cmdFabric("ssh", "-q", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", ip, cmd)
			if stdoutStderr, err := cmd.CombinedOutput(); err != nil {
				logDistant.WithField("output", fmt.Sprintf("%s", stdoutStderr)).Info("Ssh output")
				chanErr <- err
			} else {
				chanErr <- nil
			}
		}()
	}

	for i := 0; i < distant.NbDistantToLaunch; i++ {
		if err := <-chanErr; err != nil {
			errs.Errors = append(errs.Errors, err)
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (distant *amazonDistantImpl) RunAfter() error {
	err := distant.runCmd(distant.AfterCmd)
	if err != nil {
		return err
	}
	return nil
}

func (client *amazonImpl) RunBefore() error {
	runResult, err := client.ec2Client.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(client.ImageID),
		InstanceType:     aws.String(client.InstanceType),
		KeyName:          aws.String(client.KeyPairName),
		SecurityGroupIds: aws.StringSlice(client.SecurityGroupIds),
		MinCount:         aws.Int64(client.MinCount),
		MaxCount:         aws.Int64(client.MaxCount),
	})
	if err != nil {
		return err
	}

	instancesID := make([]*string, len(runResult.Instances))
	for index, instance := range runResult.Instances {
		instancesID[index] = instance.InstanceId
	}
	client.instancesID = instancesID

	logAmazon.WithField("nb", len(runResult.Instances)).Info("Created instance on amazon wait until running")
	if err := client.ec2Client.WaitUntilInstanceStatusOk(&ec2.DescribeInstanceStatusInput{InstanceIds: instancesID}); err != nil {
		return err
	}

	logAmazon.Info("Configure networks")
	for _, instance := range runResult.Instances {
		for _, network := range instance.NetworkInterfaces {
			client.ec2Client.ModifyNetworkInterfaceAttribute(&ec2.ModifyNetworkInterfaceAttributeInput{
				NetworkInterfaceId: network.NetworkInterfaceId,
				SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
			})
		}
	}
	logAmazon.WithField("nb", len(runResult.Instances)).Info("Created instance on amazon running")

	return nil
}

func (client *amazonImpl) Deploy() error {
	return nil // deploy and launch are grouped in RunAfter to not make multiple call on distant server
}

func (client *amazonImpl) RunAfter() error {
	res, err := client.ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: client.instancesID})
	if err != nil {
		return err
	}

	logAmazon.WithField("nb", len(client.instancesID)).Info("Deploy lorhammer")

	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			client.distantDeployer.IPServer = *instance.PublicDnsName
			client.distantDeployer.AfterCmd = fmt.Sprintf("nohup %s/lorhammer -mqtt %s > lorahmmer.log 2>&1 &", client.distantDeployer.PathWhereScp, client.mqttAddress)
			err := client.distantDeployer.Deploy()
			if err != nil {
				logAmazon.WithError(err).Error("Lorhammer not deployed")
			} else {
				logAmazon.Info("Lorhammer deployed")
			}

			err = client.distantDeployer.RunAfter()
			if err != nil {
				logAmazon.WithError(err).Error("Lorhammer not started")
			} else {
				logAmazon.Info("Lorhammers started")
			}
		}
	}

	logAmazon.WithField("nb", len(client.instancesID)).Info("All Lorhammers deployed !!")

	return nil
}
