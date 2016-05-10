)ckage main

import (
        "fmt"
        "log"
        "regexp"
        "sync"

        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/ec2"
)

type Pool struct {
        sync.Mutex
        Available    bool
        Vpc          string
        Filter       *regexp.Regexp
        FilterString string
        Client       *ec2.EC2
        Running      []*ec2.Instance
        Stopped      []*ec2.Instance
}

type InstanceList []*string

func (i InstanceList) String() []string {
        var list []string
        for _, instance := range i {
                list = append(list, *instance)
        }
        return list
}

// Start attempts to start n instances
// in the pools stopped list.
func (p *Pool) Start(n int) error {
        if !p.Available {
                return fmt.Errorf("[%s - %s] Pool unavailable", p.Vpc, p.FilterString)
        } else {
                log.Printf("[%s - %s] Requested start for %d instances\n", p.Vpc, p.FilterString)
        }

        var toRun InstanceList

        if len(p.Stopped) == 0 {
                return fmt.Errorf("[%s - %s] No stopped instances available", p.Vpc, p.FilterString)
        }

        if n > len(p.Stopped) {
                toRun = p.List("stopped")
                log.Printf("[%s - %s] Requested start for %d instances, only %d available\n",
                        p.Vpc, p.FilterString, n, len(p.List("stopped")))
        } else {
                toRun = p.List("stopped")[:n]
        }

        log.Printf("[%s - %s] Requesting start for %s\n", p.Vpc, p.FilterString, toRun.String())

        _, err := p.Client.StartInstances(&ec2.StartInstancesInput{
                InstanceIds: toRun,
        })
        if err != nil {
                return err
        }

        err = p.Client.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
                InstanceIds: toRun,
        })
        if err != nil {
                return err
        }

        p.Update()

        return nil
}

// Update refreshes the list of running and stopped
// instances for the pool.
func (p *Pool) Update() {
        p.Lock()

        resp, err := p.Client.DescribeInstances(&ec2.DescribeInstancesInput{
                MaxResults: aws.Int64(1000),
                Filters:    []*ec2.Filter{&ec2.Filter{Name: aws.String("vpc-id"), Values: []*string{aws.String(p.Vpc)}}},
        })
        if err != nil {
                log.Println(err)
                return
        }

        p.Running, p.Stopped = []*ec2.Instance{}, []*ec2.Instance{}
        instances := []*ec2.Instance{}
        // "We need to go deeper" -- Inception
        for _, reservation := range resp.Reservations {
                for _, instance := range reservation.Instances {
                        for _, tag := range instance.Tags {
                                if *tag.Key == "Name" {
                                        if p.Filter.Match([]byte(*tag.Value)) {
                                                instances = append(instances, instance)
                                        }
                                }
                        }
                }
        }

        for _, instance := range instances {
                switch *instance.State.Name {
                case "running":
                        p.Running = append(p.Running, instance)
                case "stopped":
                        p.Stopped = append(p.Stopped, instance)
                }
        }

        if len(p.Running) > 0 || len(p.Stopped) > 0 {
                p.Available = true
        }

        p.Unlock()

        log.Printf("[%s - %s] Pool updated - Running: %s - Stopped: %s\n",
                p.Vpc, p.FilterString, p.ListString("running"), p.ListString("stopped"))
}

// ListSTring returns a slice of InstanceID strings.
func (p *Pool) ListString(state string) []string {
        p.Lock()
        defer p.Unlock()

        ids := []string{}

        switch state {
        case "running":
                for _, i := range p.Running {
                        ids = append(ids, *i.InstanceId)
                }
        case "stopped":
                for _, i := range p.Stopped {
                        ids = append(ids, *i.InstanceId)
                }
        }

        return ids
}

// List returns a slice of InstanceIDs that are
// either running or stopped, based on the state parameter.
func (p *Pool) List(state string) []*string {
        p.Lock()
        defer p.Unlock()

        ids := []*string{}

        switch state {
        case "running":
                for _, i := range p.Running {
                        ids = append(ids, i.InstanceId)
                }
        case "stopped":
                for _, i := range p.Stopped {
                        ids = append(ids, i.InstanceId)
                }
        }

        return ids
}

// NewPool constructs a list of running and stopped
// machines of a particular class (filtered by tag.Name)
// belonging to the specified VPC.
func NewPool(vpc, filter string) *Pool {
        svc := ec2.New(session.New(&aws.Config{Region: aws.String("us-west-2")}))
        pool := &Pool{
                Available:    false,
                Client:       svc,
                Vpc:          vpc,
                FilterString: filter,
                Filter:       regexp.MustCompile(filter),
        }

        log.Printf("[%s - %s] Pool created\n", vpc, filter)
        pool.Update()

        return pool
}

func main() {
        pool := NewPool("vpc-xxx", "web")

        err := pool.Start(2)
        if err != nil {
                log.Println(err)
        }

}
