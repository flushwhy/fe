package build

import (
 "fmt"
 "os"
 "os/exec"
 "strings"
 "github.com/flushwhy/fe/internal/config"
)

type Runner func(command string,args ...string) error
var DefaultRunner Runner=func(command string,args ...string)error{
 cmd:=exec.Command(command,args...);cmd.Stdout=os.Stdout;cmd.Stderr=os.Stderr;return cmd.Run()
}
type Options struct{ Platform string }

// Run plans every selected target and checks all tools before execution.
func Run(cfg config.BuildConfig,opts Options,runner Runner)error{
 plan,err:=Plan(cfg,opts);if err!=nil{return err}
 if err:=CheckTools(plan);err!=nil{return err}
 fmt.Printf("Building %d target(s)...\n",len(plan))
 for _,target:=range plan{if err:=RunPlannedTarget(target,runner);err!=nil{return err}}
 fmt.Printf("✓ Build complete\n");return nil
}
func RunTarget(target config.BuildTarget,runner Runner)error{
 plan,err:=Plan(config.BuildConfig{Targets:[]config.BuildTarget{target}},Options{});if err!=nil{return err}
 if err:=CheckTools(plan);err!=nil{return err}
 return RunPlannedTarget(plan[0],runner)
}
func FilterByPlatform(targets []config.BuildTarget,platform string)[]config.BuildTarget{
 var out []config.BuildTarget
 for _,target:=range targets{if strings.EqualFold(target.Platform,platform){out=append(out,target)}}
 return out
}
